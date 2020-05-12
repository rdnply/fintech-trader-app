package trade

import (
	"context"
	"cw1/internal/robot"
	pb "cw1/internal/streamer"
	"cw1/pkg/log/logger"
	"fmt"
	"io"
)

type Ticker struct {
	name      string
	robots    []*robot.Robot
	clients   map[*Client]bool
	//send      chan []*robot.Robot
	start     chan []*robot.Robot
	stop      chan bool
	stopDeals chan bool
	broadcast chan []*robot.Robot
	id        map[int64]*Client
}

func (t *Ticker) run() {
	fmt.Println("start run for ticker")
	for {
		select {
		case robots := <-t.start:
			fmt.Println("creating ticker")
			for _, r := range robots {
				client := &Client{t, r, make(chan *robot.Robot)}
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
		case robots := <-t.broadcast:
			for _, r := range robots {
				if _, ok := t.id[r.RobotID]; ok {
					t.id[r.RobotID].send <- r
				}
			}

			//for client := range t.clients {
			//	select {
			//	case client.send <- message:
			//	default:
			//		close(client.send)
			//		delete(t.clients, client)
			//	}
			//}
		}
	}
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
			if err == io.EOF {
				break
			}
			if err != nil {
				l.Errorf("can't get price from request: %v", err)
			}

			fmt.Printf("Ticker: %v; %v, %v, %v\n", t.name, lot.SellPrice, lot.BuyPrice, lot.Ts)
		}
	}
}
