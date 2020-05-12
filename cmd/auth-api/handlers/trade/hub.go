package trade

import (
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
	//robots     chan []*robot.Robot
	logger     logger.Logger
	names      map[string]*Ticker
}

func NewHub(s pb.TradingServiceClient, l logger.Logger) *Hub {
	return &Hub{
		service:    s,
		tickers:    make(map[*Ticker]bool),
		broadcast:  make(chan *trade),
		register:   make(chan *Ticker),
		unregister: make(chan *Ticker),
		logger:     l,
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
			fmt.Println("make broadcase in hub for ticker...")
			if _, ok := h.names[trade.ticker]; ok {
				h.names[trade.ticker].broadcast <- trade.robots
			}
		}
	}

}

//func (h *Hub) Broadcast(rbt *robot.Robot) {
//	done := make(chan bool)
//
//	go func() {
//		h.broadcast <- rbt
//		done <- true
//	}()
//
//	<-done
//	close(done)
//}
