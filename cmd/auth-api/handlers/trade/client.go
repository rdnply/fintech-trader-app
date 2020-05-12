package trade

import (
	"cw1/internal/robot"
	"fmt"
)

type Client struct {
	ticker  *Ticker
	robot *robot.Robot
	send chan *robot.Robot
}

func (c *Client) work() {
	defer func() {
		close(c.send)
	}()
	fmt.Println("start work for client")
	for {

		select {
		case r := <-c.send:
			fmt.Printf("Robot id storage: %v\n", c.robot.RobotID)
			fmt.Printf("Make work with id id: %v\n", r.RobotID)
		}
	}
}