package websocket

import (
	"cw1/cmd/auth-api/httperror"
	"cw1/internal/robot"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan *robot.Robot
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			fmt.Printf("get message and try write json: %v\n", message.RobotID)
			err := c.conn.WriteJSON(message)
			if err != nil {
				fmt.Printf("can't write json: %v\n", err)
				return
			}
		case <-ticker.C:
			fmt.Println("ping")
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) error {
	fmt.Println("start websocket")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ctx := fmt.Sprintf("Can't open websocket connection\n")
		return httperror.NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}
	client := &Client{hub: hub, conn: conn, send: make(chan *robot.Robot)}
	client.hub.register <- client

	go client.writePump()

	return nil
}

//func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) error {
//	conn, err := upgrader.Upgrade(w, r, nil)
//	if err != nil {
//		ctx := fmt.Sprintf("Can't open websocket connection\n")
//		return httperror.NewHTTPError(ctx, err, "", http.StatusInternalServerError)
//	}
//	client := &Client{hub: hub, conn: conn, send: make(chan []byte)}
//	client.hub.register <- client
//
//	go client.writePump()
//
//	return nil
//}
