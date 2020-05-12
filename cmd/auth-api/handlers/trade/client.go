package trade

import (
	"cw1/cmd/auth-api/handlers/socket"
	"cw1/internal/postgres"
	"cw1/internal/robot"
	pb "cw1/internal/streamer"
	"cw1/pkg/log/logger"
)

type Client struct {
	ticker       *Ticker
	r            *robot.Robot
	send         chan *pb.PriceResponse
	isBuying     bool
	isSelling    bool
	robotStorage *postgres.RobotStorage
	ws           *socket.Hub
	logger       logger.Logger
}

func (c *Client) work() {
	defer func() {
		close(c.send)
	}()

	c.logger.Infof("Start client with id: %v", c.r.RobotID)

	for lot := range c.send {
		c.canMakeTrade(lot)
	}
}

func (c *Client) canMakeTrade(resp *pb.PriceResponse) {
	if !isValid(c.r) {
		return
	}

	if c.isBuying && c.r.BuyPrice.V.Float64 >= resp.BuyPrice {
		c.r.BuyPrice.V.Float64 = resp.BuyPrice
		c.isBuying = false
		c.isSelling = true
		c.logger.Infof("Buy lot: %v", resp)
	}

	if c.isSelling && c.r.SellPrice.V.Float64 <= resp.SellPrice {
		c.r.SellPrice.V.Float64 = resp.SellPrice
		c.isSelling = false
		c.logger.Infof("Sell lot: %v", resp)
	}

	if !c.isSelling && !c.isBuying {
		c.r.FactYield.V.Float64 += c.r.SellPrice.V.Float64 - c.r.BuyPrice.V.Float64
		c.r.DealsCount.V.Int64++

		err := c.robotStorage.Update(c.r)
		if err != nil {
			c.logger.Errorf("Can't update robot with id: %v", c.r.RobotID)
		}

		c.ws.Broadcast(c.r)
		c.logger.Infof("Make update for robot with id: %v", c.r.RobotID)
		c.isBuying = true
	}
}

func isValid(r *robot.Robot) bool {
	if r.BuyPrice == nil || r.SellPrice == nil || r.DealsCount == nil || r.FactYield == nil {
		return false
	}

	return true
}
