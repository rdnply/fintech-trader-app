package trade

import (
	"cw1/cmd/socket"
	"cw1/internal/robot"
	pb "cw1/internal/streamer"
	"cw1/pkg/log/logger"
	"time"
)

type Trader struct {
	tickers        map[string]bool
	tradingService pb.TradingServiceClient
	robotStorage   robot.Storage
	hub            *Hub
	ws             *socket.Hub
	logger         logger.Logger
}

type tradeInfo struct {
	name   string
	robots []*robot.Robot
}

func New(l logger.Logger, tc pb.TradingServiceClient, rs robot.Storage, ws *socket.Hub) *Trader {
	return &Trader{
		tickers:        make(map[string]bool),
		tradingService: tc,
		robotStorage:   rs,
		hub:            NewHub(tc, l, rs),
		ws:             ws,
		logger:         l,
	}
}

func (t *Trader) StartDeals(quit chan bool) {
	const Timeout = 3

	tick := time.NewTicker(time.Second * Timeout)

	go t.hub.Run()

	go func() {
		for {
			select {
			case <-tick.C:
				rbts, err := t.robotStorage.GetActiveRobots()
				if err != nil {
					t.logger.Errorf("can't get active robots from storage: %v", err)
				}

				rbtsByTicker := getRobotsByTicker(rbts)
				t.work(rbtsByTicker)
			case <-quit:
				t.logger.Infof("Quit from trader")
				tick.Stop()

				return
			}
		}
	}()
}

func (t *Trader) work(rbtsByTicker map[string][]*robot.Robot) {
	toDelete := make(map[string]bool)

	done := make(chan bool)

	go func() {
		for k := range t.tickers {
			toDelete[k] = true
		}

		for name, rbts := range rbtsByTicker {
			if !t.tickers[name] {
				ticker := initTicker(name, rbts, t.robotStorage, t.ws, t.logger, t.tradingService)
				t.tickers[name] = true
				t.hub.register <- ticker
			}

			toDelete[name] = false
		}

		for name, del := range toDelete {
			if del {
				t.hub.unregister <- name
				delete(t.tickers, name)
			}
		}

		for name, del := range toDelete {
			if !del {
				trade := &tradeInfo{name, rbtsByTicker[name]}
				t.hub.broadcast <- trade
			}
		}

		done <- true
	}()

	<-done
}

func initTicker(n string, rr []*robot.Robot, rs robot.Storage, ws *socket.Hub, l logger.Logger, s pb.TradingServiceClient) *Ticker {
	t := &Ticker{
		clients:      make(map[*Client]bool),
		ids:          make(map[int64]*Client),
		name:         n,
		robots:       rr,
		service:      s,
		robotStorage: rs,
		ws:           ws,
		start:        make(chan bool),
		stop:         make(chan bool),
		broadcast:    make(chan []*robot.Robot),
		logger:       l,
	}

	return t
}

func getRobotsByTicker(rr []*robot.Robot) map[string][]*robot.Robot {
	res := make(map[string][]*robot.Robot)
	robots := make(map[int64]bool)

	for _, r := range rr {
		id := r.RobotID
		if !robots[id] {
			res[r.Ticker.V.String] = append(res[r.Ticker.V.String], r)
			robots[id] = true
		}
	}

	return res
}
