package trade

import (
	"cw1/cmd/socket"
	"cw1/internal/robot"
	pb "cw1/internal/streamer"
	"cw1/pkg/log/logger"
)

type Client struct {
	r            *robot.Robot
	tickerName   string
	robotStorage robot.Storage
	ws           *socket.Hub
	send         chan *pb.PriceResponse
	unregister   chan bool
	isBuying     bool
	isSelling    bool
	buyPrice     float64
	sellPrice    float64
	logger       logger.Logger
}

func (c *Client) work() {
	defer func() {
		close(c.send)
		close(c.unregister)
	}()

	c.logger.Infof("Start client for robot with id: %v", c.r.RobotID)

	for {
		select {
		case lot := <-c.send:
			c.makeTrade(lot)
		case <-c.unregister:
			c.logger.Infof("Stop client for robot with id: %v", c.r.RobotID)
			return
		}
	}
}

func (c *Client) makeTrade(resp *pb.PriceResponse) {
	if !isValid(c.r) {
		return
	}

	if c.isBuying && c.r.BuyPrice.V.Float64 >= resp.BuyPrice {
		c.buyPrice = resp.BuyPrice
		c.isBuying = false
		c.isSelling = true
		c.logger.Infof("Buy %v lot with price: buy price:%v, sell price: %v; border for buy: %v",
			c.tickerName, resp.BuyPrice, resp.SellPrice, c.r.BuyPrice.V.Float64)
	}

	if c.isSelling && c.r.SellPrice.V.Float64 <= resp.SellPrice {
		c.sellPrice = resp.SellPrice
		c.isSelling = false
		c.logger.Infof("Sell %v lot with price: buy price:%v, sell price: %v; border for sell: %v",
			c.tickerName, resp.BuyPrice, resp.SellPrice, c.r.SellPrice.V.Float64)
	}

	if !c.isSelling && !c.isBuying {
		c.r.FactYield.V.Float64 += c.sellPrice - c.buyPrice
		c.r.DealsCount.V.Int64++

		err := c.robotStorage.UpdateBesidesActive(c.r)
		if err != nil {
			c.logger.Errorf("can't update robot with ids: %v", c.r.RobotID)
		}

		c.ws.Broadcast(c.r)
		c.isBuying = true
	}
}

func isValid(r *robot.Robot) bool {
	if r.BuyPrice == nil || r.SellPrice == nil ||
		r.DealsCount == nil || r.FactYield == nil {
		return false
	}

	return true
}
