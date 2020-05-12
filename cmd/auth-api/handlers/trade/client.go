package trade

import (
	"cw1/cmd/auth-api/handlers/socket"
	"cw1/internal/postgres"
	"cw1/internal/robot"
	pb "cw1/internal/streamer"
	"fmt"
)

type Client struct {
	ticker *Ticker
	r      *robot.Robot
	//send chan *r.Robot
	send     chan *pb.PriceResponse
	isBuying bool
	isSelling bool
	robotStorage   *postgres.RobotStorage
	ws *socket.Hub
}

func (c *Client) work() {
	defer func() {
		close(c.send)
	}()
	fmt.Println("start work for client")
	for {

		select {
		case lot := <-c.send:
			//updating r info
			c.canMakeTrade(lot)
			//fmt.Printf("Robot id storage: %v\n", c.r.RobotID)
			//fmt.Printf("Lot for r: %v\n", lot)

		}
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
	}
	if c.isSelling && c.r.SellPrice.V.Float64 <= resp.SellPrice {
		c.r.SellPrice.V.Float64 = resp.SellPrice
		c.isSelling = false
	}
	fmt.Println("Lot:", resp)
	if !c.isSelling && !c.isBuying {
		c.r.FactYield.V.Float64 += c.r.SellPrice.V.Float64 - c.r.BuyPrice.V.Float64
		c.r.DealsCount.V.Int64++

		err := c.robotStorage.Update(c.r)
		if err != nil {
			fmt.Errorf("can't update robot for change trading info")
		}
		c.ws.Broadcast(c.r)
		fmt.Println("make update for robot id:", c.r.RobotID)
		c.isBuying = true
	}
}

func isValid(r *robot.Robot) bool {
	if r.BuyPrice == nil || r.SellPrice == nil || r.DealsCount == nil || r.FactYield == nil {
		return false
	}

	return true
}

