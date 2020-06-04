package trade

import (
	"cw1/internal/robot"
	pb "cw1/internal/streamer"
	"cw1/pkg/log/logger"
)

type Hub struct {
	tickers      map[*Ticker]bool
	names        map[string]*Ticker
	service      pb.TradingServiceClient
	robotStorage robot.Storage
	register     chan *Ticker
	unregister   chan string
	broadcast    chan *tradeInfo
	logger       logger.Logger
}

func NewHub(s pb.TradingServiceClient, l logger.Logger, rs robot.Storage) *Hub {
	return &Hub{
		tickers:      make(map[*Ticker]bool),
		names:        make(map[string]*Ticker),
		service:      s,
		robotStorage: rs,
		register:     make(chan *Ticker),
		unregister:   make(chan string),
		broadcast:    make(chan *tradeInfo),
		logger:       l,
	}
}

func (h *Hub) Run() {
	h.logger.Infof("Start running hub for tickers")

	defer func() {
		close(h.register)
		close(h.unregister)
		close(h.broadcast)
	}()

	for {
		select {
		case ticker := <-h.register:
			go ticker.run()
			ticker.start <- true

			h.names[ticker.name] = ticker
			h.tickers[ticker] = true

		case name := <-h.unregister:
			if _, ok := h.names[name]; ok {
				ticker := h.names[name]
				ticker.stop <- true
				delete(h.tickers, ticker)
			}

		case trade := <-h.broadcast:
			if _, ok := h.names[trade.name]; ok {
				h.names[trade.name].broadcast <- trade.robots
			}
		}
	}
}
