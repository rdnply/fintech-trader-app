package trade

import (
	"context"
	"cw1/cmd/socket"
	"cw1/internal/robot"
	pb "cw1/internal/streamer"
	"cw1/pkg/log/logger"
	"io"
	"sync"
)

type Ticker struct {
	mu           sync.Mutex
	clients      map[*Client]bool
	ids          map[int64]*Client
	name         string
	robotStorage robot.Storage
	ws           *socket.Hub
	logger       logger.Logger
	robots       []*robot.Robot
	start        chan bool
	stop         chan bool
	broadcast    chan []*robot.Robot
	service      pb.TradingServiceClient
}

func (t *Ticker) run() {
	defer func() {
		close(t.start)
		close(t.stop)
		close(t.broadcast)
	}()

	for {
		select {
		case <-t.start:
			t.logger.Infof("Start ticker with name: %v", t.name)

			go t.makeDeals()

			for _, r := range t.robots {
				client := initClient(t, r, t.robotStorage, t.ws, t.logger)
				t.ids[r.RobotID] = client
				t.mu.Lock()
				t.clients[client] = true
				t.mu.Unlock()

				go client.work()
			}
		case <-t.stop:
			t.logger.Infof("Stop ticker with name: %v", t.name)
			t.mu.Lock()
			for c := range t.clients {
				t.logger.Infof("Close client with ID: %v", c.r.RobotID)
				c.unregister <- true
				delete(t.clients, c)
			}

			t.mu.Unlock()

			return
		case robots := <-t.broadcast:
			toWork := t.workWithRobots(robots)
			for _, client := range toWork {
				go client.work()
			}
		}
	}
}

func (t *Ticker) workWithRobots(rbts []*robot.Robot) []*Client {
	toDelete := make(map[int64]bool)
	toWork := make([]*Client, 0)

	done := make(chan bool)

	go func() {
		for k := range t.clients {
			toDelete[k.r.RobotID] = true
		}

		for _, r := range rbts {
			if _, ok := t.ids[r.RobotID]; !ok {
				t.logger.Infof("Register client with id: %v", r.RobotID)
				client := initClient(t, r, t.robotStorage, t.ws, t.logger)
				t.mu.Lock()
				t.clients[client] = true
				t.mu.Unlock()
				t.ids[r.RobotID] = client
				toWork = append(toWork, client)
			} else {
				t.mu.Lock()
				client := t.ids[r.RobotID]
				client.r = r
				t.mu.Unlock()
			}

			toDelete[r.RobotID] = false
		}

		for id, del := range toDelete {
			if del {
				t.logger.Infof("Delete client with id: %v", id)
				t.ids[id].unregister <- true
				t.mu.Lock()
				delete(t.clients, t.ids[id])
				delete(t.ids, id)
				t.mu.Unlock()
			}
		}

		done <- true
	}()

	<-done

	return toWork
}

func initClient(t *Ticker, r *robot.Robot, rs robot.Storage, ws *socket.Hub, l logger.Logger) *Client {
	c := &Client{
		ticker:       t,
		r:            r,
		send:         make(chan *pb.PriceResponse),
		isBuying:     true,
		isSelling:    false,
		robotStorage: rs,
		ws:           ws,
		logger:       l,
		unregister:   make(chan bool),
	}

	return c
}

func (t *Ticker) makeDeals() {
	priceRequest := pb.PriceRequest{Ticker: t.name}

	resp, err := t.service.Price(context.Background(), &priceRequest)
	if err != nil {
		t.logger.Errorf("can't get prices from stream: %v", err)
		return
	}

	for {
		lot, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			t.logger.Errorf("can't get price from request: %v", err)
			return
		}

		t.mu.Lock()
		for c := range t.clients {
			c.send <- lot
		}
		t.mu.Unlock()
	}
}
