package trade

import (
	"cw1/cmd/auth-api/handlers/socket"
	"cw1/internal/postgres"
	"cw1/internal/robot"
	pb "cw1/internal/streamer"
	"cw1/pkg/log/logger"
	"fmt"
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
	ticker := time.NewTicker(time.Second * 15)
	t.logger.Infof("Start trade")
	go t.hub.Run()

	go func() {
		for {
			select {
			case <-ticker.C:
				rbts, err := t.robotStorage.GetActiveRobots()
				if err != nil {
					t.logger.Errorf("Can't get active robots from storage: %v", err)
				}

				fmt.Println(rbts)
				rbtsByTicker := getRobotsByTicker(rbts)

				toDelete := make(map[string]bool)

				for k := range t.tickers {
					toDelete[k] = true
				}

				for k, v := range rbtsByTicker {
					if !t.tickers[k] {
						ticker := initTicker(k, v, t.robotStorage, t.ws)
						t.hub.register <- ticker
					}
					toDelete[k] = false
				}

				for k, del := range toDelete {
					if del {
						ticker := initTicker(k, nil, t.robotStorage, t.ws)
						t.hub.unregister <- ticker
					}
				}

				for k, del := range toDelete {
					if !del {
						trade := &trade{k, rbtsByTicker[k]}
						t.hub.broadcast <- trade
					}
				}
			case <-quit:
				fmt.Println("quit from trader")
				ticker.Stop()
				return
			}
		}
	}()

}

func initTicker(n string, rr []*robot.Robot, rs *postgres.RobotStorage, ws *socket.Hub) *Ticker {
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
		ws: ws,
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
