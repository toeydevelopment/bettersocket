package bettersocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// func (s subscription) readPump() {
// 	c := s.conn
// 	defer func() {
// 		h.unregister <- s
// 		c.ws.Close()
// 	}()

// 	c.ws.SetReadLimit(maxMessageSize)
// 	c.ws.SetReadDeadline(time.Now().Add(pongWait))
// 	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
// 	for {
// 		_, msg, err := c.ws.ReadMessage()
// 		if err != nil {
// 			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
// 				log.Printf("error: %v", err)
// 			}
// 			break
// 		}
// 		m := message{msg, s.room}
// 		Hub.broadcast <- m
// 	}
// }
