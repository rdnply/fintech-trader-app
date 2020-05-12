package trade

import (
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
}

type trade struct {
	ticker string
	robots []*robot.Robot
}

func New(l logger.Logger, tc pb.TradingServiceClient, rs *postgres.RobotStorage) *Trader {
	return &Trader{
		logger:         l,
		tradingService: tc,
		robotStorage:   rs,
		hub:            NewHub(tc, l),
		tickers:        make(map[string]bool),
	}
}

func (t *Trader) StartDeals(quit chan bool) {
	ticker := time.NewTicker(time.Second * 10)
	t.logger.Infof("Start trade")
	go t.hub.Run()

	go func() {
		for {
			select {
			case <-ticker.C:
				rbts, err := t.robotStorage.GetActiveRobots()
				if err != nil {
					t.logger.Errorf("Can't get active robots from storage: %del", err)
				}

				fmt.Println(rbts)
				rbtsByTicker := getRobotsByTicker(rbts)

				toDelete := make(map[string]bool)

				for k := range t.tickers {
					toDelete[k] = true
				}


				for k, v := range rbtsByTicker {
					if !t.tickers[k] {
						ticker := initTicker(k, v)
						t.hub.register <- ticker
					}
					toDelete[k] = false
				}

				for k, del := range toDelete {
					if del {
						ticker := initTicker(k, nil)
						t.hub.unregister <- ticker
					}
				}

				for k, del := range toDelete {
					if !del {
						trade := &trade{k,  rbtsByTicker[k]}
						t.hub.broadcast <- trade
					}
				}


				//var wg sync.WaitGroup
				//for k, del := range rbtsByTicker {
				//	//wg.Add(1)
				//	makeDeals(&wg, k, del, t.tradingService, t.logger)
				//}
				////wg.Wait()
				//fmt.Println(rbtsByTicker)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

}

func initTicker(n string , rr []*robot.Robot) *Ticker {
	t := &Ticker{
		name: n,
		robots: rr,
		clients: make(map[*Client]bool),
		start: make(chan []*robot.Robot),
		stop: make(chan bool),
		stopDeals: make(chan bool),
		broadcast: make(chan []*robot.Robot),
		id: make(map[int64]*Client),
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

//func makeDeals(wg *sync.WaitGroup, ticker string, rr []int64, service pb.TradingServiceClient, l logger.Logger) {
//	priceRequest := pb.PriceRequest{Ticker: ticker}
//	resp, err := service.Price(context.Background(), &priceRequest)
//	if err != nil {
//		l.Errorf("can't get prices from stream: %v", err)
//		return
//	}
//	fmt.Println("get prices")
//	for {
//		lot, err := resp.Recv()
//		if err == io.EOF {
//			wg.Done()
//			break
//		}
//		if err != nil {
//			l.Errorf("can't get price from request: %v", err)
//		}
//
//		fmt.Printf("Ticker: %v; %v, %v, %v\n", ticker, lot.SellPrice, lot.BuyPrice, lot.Ts)
//	}
//
//}
