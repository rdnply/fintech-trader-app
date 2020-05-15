package trade

import (
	"context"
	"cw1/cmd/socket"
	"cw1/internal/robot"
	pb "cw1/internal/streamer"
	"cw1/pkg/log/logger"
	"io"
)

type Ticker struct {
	name         string
	robots       []*robot.Robot
	clients      map[*Client]bool
	start        chan []*robot.Robot
	stop         chan bool
	stopDeals    chan bool
	broadcast    chan []*robot.Robot
	id           map[int64]*Client
	robotStorage robot.Storage
	ws           *socket.Hub
	logger       logger.Logger
}

func (t *Ticker) run() {
	//t.logger.Infof("Start ticker with name: %v", t.name)
	for {
		select {
		case robots := <-t.start:
			for _, r := range robots {
				client := initClient(t, r, t.robotStorage, t.ws, t.logger)
				t.clients[client] = true

				client.work()
			}
		case <-t.stop:
			t.stopDeals <- true
			for c := range t.clients {
				delete(t.clients, c)
				close(c.send)
			}

			return
		}
	}
}

func initClient(t *Ticker, r *robot.Robot, rs robot.Storage, ws *socket.Hub, l logger.Logger) *Client {
	c := &Client{
		ticker:       t,
		r:            r,
		send:         make(chan *pb.PriceResponse),
		isBuying:     false,
		isSelling:    false,
		robotStorage: rs,
		ws:           ws,
		logger:       l,
	}

	return c
}

func (t *Ticker) makeDeals(service pb.TradingServiceClient, l logger.Logger) {
	//t.logger.Infof("Start making deals for ticker with name: %v", t.name)
	priceRequest := pb.PriceRequest{Ticker: t.name}

	resp, err := service.Price(context.Background(), &priceRequest)
	if err != nil {
		l.Errorf("can't get prices from stream: %v", err)
		return
	}

	for {
		lot, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			l.Errorf("can't get price from request: %v", err)
			return
		}

		for c := range t.clients {
			c.send <- lot
		}
	}
}
