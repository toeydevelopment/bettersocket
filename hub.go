package bettersocket

import (
	"context"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type subscription struct {
	conn *connection
	room string
}

type message struct {
	data []byte
	room string
}

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	// Registered connections.
	rooms map[string]map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan message

	// Register requests from the connections.
	register chan subscription

	// Unregister requests from connections.
	unregister chan subscription
}

func (h *hub) run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case s := <-h.register:

			connections := h.rooms[s.room]

			if connections == nil {
				// initate connection (first time ??)
				connections = make(map[*connection]bool)

				h.rooms[s.room] = connections
			}

			h.rooms[s.room][s.conn] = true
			continue
			// end register
		case us := <-h.unregister:
			connections := h.rooms[us.room]

			if connections != nil {
				if _, existed := connections[us.conn]; existed {
					delete(connections, us.conn)
					close(us.conn.send)
					if len(connections) == 0 {
						delete(h.rooms, us.room)
					}
				}
			}

			continue
			// end unregister
		case broadcast := <-h.broadcast:
			connections := h.rooms[broadcast.room]
			for c := range connections {
				select {
				case c.send <- broadcast.data:
				default:
					close(c.send)
					delete(connections, c)
					if len(connections) == 0 {
						delete(h.rooms, broadcast.room)
					}
				}
			}
		}
	}
}

var h = hub{
	broadcast:  make(chan message),
	register:   make(chan subscription),
	unregister: make(chan subscription),
	rooms:      make(map[string]map[*connection]bool),
}

func (s subscription) readPump() {
	c := s.conn
	defer func() {
		c.bs.h.unregister <- s
		c.ws.Close()
	}()

	c.ws.SetReadLimit(int64(s.conn.bs.c.MaxMessageSize))
	c.ws.SetReadDeadline(time.Now().Add(s.conn.bs.c.PongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(s.conn.bs.c.PongWait))
		return nil
	})

	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}

		m := message{msg, s.room}

		s.conn.bs.h.broadcast <- m
	}
}

func (s *subscription) writePump() {
	c := s.conn
	ticker := time.NewTicker(s.conn.bs.c.pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(c.bs.c.WriteWait))
	return c.ws.WriteMessage(mt, payload)
}
