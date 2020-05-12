package trade

import (
	"cw1/internal/postgres"
	pb "cw1/internal/streamer"
	"cw1/pkg/log/logger"
	"fmt"
)

type Hub struct {
	service    pb.TradingServiceClient
	tickers    map[*Ticker]bool
	register   chan *Ticker
	unregister chan *Ticker
	broadcast  chan *trade
	//robots     chan []*r.Robot
	logger     logger.Logger
	names      map[string]*Ticker
	robotStorage   *postgres.RobotStorage
}

func NewHub(s pb.TradingServiceClient, l logger.Logger, rs *postgres.RobotStorage) *Hub {
	return &Hub{
		service:      s,
		tickers:      make(map[*Ticker]bool),
		broadcast:    make(chan *trade),
		register:     make(chan *Ticker),
		unregister:   make(chan *Ticker),
		logger:       l,
		robotStorage: rs,
	}
}

func (h *Hub) Run() {
	fmt.Println("run hub")
	for {
		fmt.Println("hub for tickers is running...")
		select {
		case ticker := <-h.register:
			fmt.Println("start register ticker")
			go ticker.run()
			go ticker.makeDeals(h.service, h.logger)
			ticker.start <- ticker.robots
			h.tickers[ticker] = true


		case ticker := <-h.unregister:
			if _, ok := h.tickers[ticker]; ok {
				ticker.stop <- true
				delete(h.tickers, ticker)
				close(ticker.start)
				close(ticker.stop)
			}
		case trade := <-h.broadcast:
			fmt.Println("make broadcast in hub for ticker...")
			if _, ok := h.names[trade.ticker]; ok {
				h.names[trade.ticker].broadcast <- trade.robots
			}
		}
	}

}

