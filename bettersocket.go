package bettersocket

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type BetterSocketer interface {
}

type BetterSocket struct {
	c        Config
	h        *hub
	mu       *sync.Mutex
	started  bool
	upgrader websocket.Upgrader
}

func New(c Config) {

	bs := new(BetterSocket)

	bs.c = c

	bs.mu = &sync.Mutex{}

	bs.upgrader = websocket.Upgrader{
		ReadBufferSize:  int(c.ReadBufferSize),
		WriteBufferSize: int(c.WriteBufferSize),
		CheckOrigin:     c.CheckOriginFunc,
	}

	bs.h = &hub{
		broadcast:  make(chan message),
		register:   make(chan subscription),
		unregister: make(chan subscription),
		rooms:      make(map[string]map[*connection]bool),
	}
}

func (bs *BetterSocket) Start(ctx context.Context) error {
	if bs.started {
		return errors.New("already start")
	}
	return bs.h.run(ctx)
}

func (bs *BetterSocket) ServeWS(ctx context.Context, w http.ResponseWriter, r *http.Request, roomID string) error {
	if !bs.started {
		return errors.New("run Start first")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	ws, err := bs.upgrader.Upgrade(w, r, nil)

	if err != nil {
		return err
	}

	c := &connection{
		ws:   ws,
		send: make(chan []byte, 256),
	}

	s := subscription{conn: c, room: roomID}

	bs.h.register <- s

	go s.readPump(bs)
	go s.writePump(bs)

	return nil

}
