package trade

import (
	"cw1/cmd/socket"
	"cw1/internal/postgres"
	"cw1/internal/robot"
	pb "cw1/internal/streamer"
	"cw1/pkg/log/logger"
	"time"
)

type Trader struct {
	logger         logger.Logger
	tradingService pb.TradingServiceClient
	robotStorage   *postgres.RobotStorage
	hub            *Hub
	tickers        map[string]bool
	ws             *socket.Hub
}

type trade struct {
	ticker string
	robots []*robot.Robot
}

func New(l logger.Logger, tc pb.TradingServiceClient, rs *postgres.RobotStorage, ws *socket.Hub) *Trader {
	return &Trader{
		logger:         l,
		tradingService: tc,
		robotStorage:   rs,
		hub:            NewHub(tc, l, rs),
		tickers:        make(map[string]bool),
		ws:             ws,
	}
}

func (t *Trader) StartDeals(quit chan bool) {
	const Timeout = 3

	ticker := time.NewTicker(time.Second * Timeout)

	//t.logger.Infof("Start making deals")

	go t.hub.Run()

	go func() {
		for {
			select {
			case <-ticker.C:
				rbts, err := t.robotStorage.GetActiveRobots()
				if err != nil {
					t.logger.Errorf("Can't get active robots from storage: %v", err)
				}

				rbtsByTicker := getRobotsByTicker(rbts)
				t.work(rbtsByTicker)
			case <-quit:
				t.logger.Infof("Quit from trader")
				ticker.Stop()

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

		for k, v := range rbtsByTicker {
			if !t.tickers[k] {
				ticker := initTicker(k, v, t.robotStorage, t.ws, t.logger)
				t.hub.register <- ticker
			}

			toDelete[k] = false
		}

		for k, del := range toDelete {
			if del {
				ticker := initTicker(k, nil, t.robotStorage, t.ws, t.logger)
				t.hub.unregister <- ticker
			}
		}

		for k, del := range toDelete {
			if !del {
				trade := &trade{k, rbtsByTicker[k]}
				t.hub.broadcast <- trade
			}
		}
		done <- true
	}()

	<-done
}

func initTicker(n string, rr []*robot.Robot, rs *postgres.RobotStorage, ws *socket.Hub, l logger.Logger) *Ticker {
	t := &Ticker{
		name:         n,
		robots:       rr,
		clients:      make(map[*Client]bool),
		start:        make(chan []*robot.Robot),
		stop:         make(chan bool),
		stopDeals:    make(chan bool),
		broadcast:    make(chan []*robot.Robot),
		id:           make(map[int64]*Client),
		robotStorage: rs,
		ws:           ws,
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
