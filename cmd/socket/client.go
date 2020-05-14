package socket

import (
	"cw1/cmd/auth-api/httperror"
	"cw1/internal/robot"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait    = 10 * time.Second
	pongWait     = 60 * time.Second
	pingPeriod   = (pongWait * 9) / 10
	readBufSize  = 1024
	writeBufSize = 1024
)

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
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteJSON(message)
			if err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) error {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  readBufSize,
		WriteBufferSize: writeBufSize,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ctx := fmt.Sprintf("Can't open socket connection\n")
		return httperror.NewHTTPError(ctx, err, "", http.StatusInternalServerError)
	}

	client := &Client{hub: hub, conn: conn, send: make(chan *robot.Robot)}
	client.hub.register <- client

	go client.writePump()

	return nil
}
