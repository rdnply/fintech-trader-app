package trade

import (
	"cw1/internal/robot"
	pb "cw1/internal/streamer"
	"cw1/pkg/log/logger"
	"fmt"
)

type Hub struct {
	service      pb.TradingServiceClient
	tickers      map[*Ticker]bool
	register     chan *Ticker
	//unregister   chan *Ticker
	unregister   chan string
	broadcast    chan *trade
	logger       logger.Logger
	names        map[string]*Ticker
	robotStorage robot.Storage
}

func NewHub(s pb.TradingServiceClient, l logger.Logger, rs robot.Storage) *Hub {
	return &Hub{
		service:      s,
		tickers:      make(map[*Ticker]bool),
		broadcast:    make(chan *trade),
		register:     make(chan *Ticker),
		unregister:   make(chan string),
		names:        make(map[string]*Ticker),
		logger:       l,
		robotStorage: rs,
	}
}

func (h *Hub) Run() {
	h.logger.Infof("Start running hub for tickers")

	for {
		select {
		case ticker := <-h.register:
			go ticker.run()
			//go ticker.makeDeals(h.service, h.logger)
			ticker.start <- true

			h.names[ticker.name] = ticker
			h.tickers[ticker] = true

		case name := <-h.unregister:
			h.logger.Infof("Remove ticker in hub with name: %v", name)

			if _, ok := h.names[name]; ok {
				ticker := h.names[name]
				ticker.stop <- true
				delete(h.tickers, ticker)
				//close(ticker.start)
				//close(ticker.stop)
			}

		case trade := <-h.broadcast:
			fmt.Println("BROADCAST IN HUB: ", trade)
			if _, ok := h.names[trade.name]; ok {
				h.names[trade.name].broadcast <- trade.robots
			}
		}
	}
}
