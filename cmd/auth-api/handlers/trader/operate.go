package trader

import (
	"context"
	"cw1/cmd/auth-api/handlers/websocket"
	"cw1/internal/postgres"
	"cw1/internal/robot"
	"cw1/internal/streamer"
	"cw1/pkg/log/logger"
	"fmt"
	"io"
	"time"
)

type Trader struct {
	logger         logger.Logger
	tradingService streamer.TradingServiceClient
	robotStorage   *postgres.RobotStorage
	hub            *websocket.Hub
}

type info struct {
	tickers        map[string]bool
	robots         map[int64]bool
	robotsByTicker map[string][]int64
}

func New(l logger.Logger, tc streamer.TradingServiceClient, rs *postgres.RobotStorage, h *websocket.Hub) *Trader {
	return &Trader{
		logger:         l,
		tradingService: tc,
		robotStorage:   rs,
		hub:            h,
	}
}

func (t *Trader) StartDeals(quit chan bool) {
	ticker := time.NewTicker(time.Second * 55)
	t.logger.Infof("Start trader")
	go func() {
		for {
			select {
			case <-ticker.C:
				rbts, err := t.robotStorage.GetActiveRobots()
				if err != nil {
					t.logger.Errorf("Can't get active robots from storage: %v", err)
				}

				rbtsByTicker := getRobotsByTicker(rbts)
				for k, v := range rbtsByTicker {
					go makeDeals(k, v, t.tradingService, t.logger)
				}

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

}


func getRobotsByTicker(rr []*robot.Robot) map[string][]int64 {
	res := make(map[string][]int64)
	robots := make(map[int64]bool)

	for _, r := range rr {
		id := r.RobotID
		if !robots[id] {
			res[r.Ticker.V.String] = append(res[r.Ticker.V.String], id)
			robots[id] = true
		}
	}

	return res
}


func makeDeals(ticker string, rr []int64, service streamer.TradingServiceClient, l logger.Logger) {
	priceRequest := streamer.PriceRequest{Ticker: ticker}
	resp, err := service.Price(context.Background(), &priceRequest)
	if err != nil {
		l.Errorf("can't get prices from stream: %v", err)
	}

	for {
		lot, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			l.Errorf("can't get price from request: %v" , err)
		}
		fmt.Printf("Ticker: %v; %v, %v, %v", ticker, lot.SellPrice, lot.BuyPrice, lot.Ts)
	}

}