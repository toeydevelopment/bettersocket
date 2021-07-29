package bettersocket

import (
	"net/http"
	"time"
)

type Config struct {
	WriteWait time.Duration

	PongWait time.Duration

	pingPeriod time.Duration

	MaxMessageSize uint

	ReadBufferSize uint

	WriteBufferSize uint

	CheckOriginFunc func(r *http.Request) bool
}

func NewConfig(c Config) *Config {

	if c.MaxMessageSize == 0 {
		c.MaxMessageSize = 1024
	}
	if c.ReadBufferSize == 0 {
		c.ReadBufferSize = 1024
	}
	if c.WriteBufferSize == 0 {
		c.WriteBufferSize = 1024
	}

	if c.CheckOriginFunc == nil {
		c.CheckOriginFunc = func(r *http.Request) bool {
			return true
		}
	}

	c.pingPeriod = (c.PongWait * 9) / 10

	return &c
}
