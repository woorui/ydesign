package core

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

const (
	// ClientTypeNone is connection type "None".
	ClientTypeNone ClientType = 0xFF
	// ClientTypeSource is connection type "Source".
	ClientTypeSource ClientType = 0x5F
	// ClientTypeUpstreamZipper is connection type "Upstream Zipper".
	ClientTypeUpstreamZipper ClientType = 0x5E
	// ClientTypeStreamFunction is connection type "Stream Function".
	ClientTypeStreamFunction ClientType = 0x5D
)

// ClientType represents the connection type.
type ClientType byte

func (c ClientType) String() string {
	switch c {
	case ClientTypeSource:
		return "Source"
	case ClientTypeUpstreamZipper:
		return "Upstream Zipper"
	case ClientTypeStreamFunction:
		return "Stream Function"
	default:
		return "None"
	}
}

type Client struct {
	ctx context.Context

	addr string
	// TLSConfig provides a TLS configuration for use by server. It must be
	// set for ListenAndServe and Serve methods.
	TLSConfig *tls.Config

	// QuicConfig provides the parameters for QUIC connection created with
	// Serve. If nil, it uses reasonable default values.
	QuicConfig *quic.Config

	cond *sync.Cond

	connected bool

	conn quic.Connection

	stream quic.Stream
}

func NewClient(ctx context.Context, addr string, tlsConfig *tls.Config, quicConfig *quic.Config) *Client {
	client := &Client{
		ctx:        context.Background(),
		cond:       sync.NewCond(&sync.Mutex{}),
		TLSConfig:  tlsConfig,
		QuicConfig: quicConfig,
	}

	client.connect(ctx, addr)

	go client.reconnect()

	return client
}

func (c *Client) connect(ctx context.Context, addr string) error {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	conn, err := quic.DialAddrContext(ctx, addr, c.TLSConfig, c.QuicConfig)
	if err != nil {
		return err
	}
	c.conn = conn

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return err
	}
	c.stream = stream

	c.connected = true

	return nil
}

func (c *Client) reconnect() {
	// TODO: add more strategy support
	maxTimes := 5
	for n := 0; n < maxTimes; n++ {
		err := c.connect(c.ctx, c.addr)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		c.cond.Broadcast()
		break
	}
}

// func (c *Client) WriteFrame(ctx context.Context, frame Frame) error {
// 	c.cond.L.Lock()
// 	defer c.cond.L.Unlock()

// 	for !c.connected {
// 		select {
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		default:
// 			c.cond.Wait()
// 		}
// 	}

// 	_, err := c.stream.Write(frame.Encode())
// 	if err != nil {
// 		c.connected = false
// 		go c.reconnect()
// 	}

// 	return err
// }
