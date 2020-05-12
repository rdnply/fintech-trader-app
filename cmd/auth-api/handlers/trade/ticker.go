package trade

import (
	"context"
	"cw1/internal/postgres"
	"cw1/internal/robot"
	pb "cw1/internal/streamer"
	"cw1/pkg/log/logger"
	"fmt"
	"io"
)

type Ticker struct {
	name    string
	robots  []*robot.Robot
	clients map[*Client]bool
	//send      chan []*r.Robot
	start        chan []*robot.Robot
	stop         chan bool
	stopDeals    chan bool
	broadcast    chan []*robot.Robot
	id           map[int64]*Client
	robotStorage *postgres.RobotStorage
}

func (t *Ticker) run() {
	fmt.Println("start run for ticker")
	for {
		select {
		case robots := <-t.start:
			fmt.Println("creating ticker")
			for _, r := range robots {
				client := initClient(t, r, t.robotStorage)
					t.clients[client] = true
				client.work()
			}
		case <-t.stop:
			t.stopDeals <- true
			for c, _ := range t.clients {
				delete(t.clients, c)
				close(c.send)
			}
			return
			//case robots := <-t.broadcast:
			//	for _, r := range robots {
			//		if _, ok := t.id[r.RobotID]; ok {
			//			t.id[r.RobotID].send <- r
			//		}
			//	}
		}
	}
}

func initClient(t *Ticker, r *robot.Robot, rs *postgres.RobotStorage) *Client {
	c := &Client{
		ticker:t,
		r: r,
		send: make(chan *pb.PriceResponse),
		isBuying: false,
		isSelling: false,
		robotStorage: rs,
	}

	return c
}

func (t *Ticker) makeDeals(service pb.TradingServiceClient, l logger.Logger) {
	fmt.Println("start making deals...")
	priceRequest := pb.PriceRequest{Ticker: t.name}
	resp, err := service.Price(context.Background(), &priceRequest)
	if err != nil {
		l.Errorf("can't get prices from stream: %v", err)
		return
	}
	fmt.Println("get prices")

	defer close(t.stopDeals)

	for {
		select {
		case <-t.stopDeals:
			return
		default:
			lot, err := resp.Recv()
			//fmt.Println("Ticker:", lot)
			//if lot == nil {
			//	continue
			//}
			if err == io.EOF {
				break
			}
			if err != nil {
				l.Errorf("can't get price from request: %v", err)
			}
			for c := range t.clients {
				c.send <- lot
			}

			//fmt.Printf("Ticker: %v; %v, %v, %v\n", t.name, lot.SellPrice, lot.BuyPrice, lot.Ts)
		}
	}
}
