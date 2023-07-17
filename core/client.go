package core

import (
	"context"
	"errors"
	"io"

	"github.com/quic-go/quic-go"
	"github.com/woorui/ydesign/core/frame"
	"golang.org/x/exp/slog"
)

// Client represents a peer in a network that can open writers and observe other writers,
// and handle them in an consumer.
type Client struct {
	conn   UniStreamPeerConnection
	option *ClientOption
}

// ClientOption is the option to create a client.
type ClientOption struct {
	Codec       frame.Codec
	PacketCodec frame.PacketCodec
	logger      *slog.Logger
	IDGenerator func() string
}

func initOption(o *ClientOption) *ClientOption {
	if o == nil {
		// fill option with defaults.
		return o
	}

	return o
}

// NewClient returns a new Client from a connection.
func NewClient(conn UniStreamPeerConnection, option *ClientOption) *Client {
	option = initOption(option)

	client := &Client{
		conn:   conn,
		option: option,
	}

	return client
}

func OpenClient(ctx context.Context, addr string) (*Client, error) {
	qconn, err := quic.DialAddr(ctx, addr, nil, nil)
	if err != nil {
		return nil, err
	}

	stream0, err := qconn.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}

}

type baseConnection struct {
	conn    quic.Connection
	stream0 quic.Stream
}

type Writer struct {
	tag        string
	opener     WriterOpener
	wrCh       chan []byte
	rdCh       chan int
	openrCh    chan WriterOpener
	underlying io.WriteCloser
}

type WriterOpenerOpener interface {
}

func MustOpenWriter(opener WriterOpener, tag string) io.Writer {
	for {
		select {
		case <-opener.Context().Done():

		}
	}
}

func (w *Writer) Write(p []byte) (int, error) {
	var (
		opener = w.opener
		writer = w.underlying
		err    error
	)

	select {
	case opener := <-w.openrCh:
		writer, err = opener.Open(w.tag)
		if err == errors.New("client: be rejected") {
			return 0, err
		}
	case p := <-w.wrCh:

	}

	io.Pipe()

	return w.underlying.Write(p)
}

func (w *Writer) beRejected(err error) bool {
	return false
}

// Open opens a writer with the given tag, which other peers can observe.
// The returned writer can be used to write to the stream associated with the given tag.
func (c *Client) Open(tag string) (io.WriteCloser, error) {
	w, err := c.conn.OpenUniStream()
	if err != nil {
		return nil, err
	}

	c.option.logger.Debug("peer opens a writer", "tag", tag)

	id := c.option.IDGenerator()

	return w, c.fillWriterFunc(id, tag, w)
}

// Observe observes tagged streams and handles them in an observer.
// The observer is responsible for handling the tagged streams and writing to a new peer stream.
func (c *Client) Observe(tag string, observer Observer) error {
	// peer request to observe stream in the specified tag.
	err := p.conn.RequestObserve(tag)
	if err != nil {
		return err
	}
	// then waiting and handling the stream reponsed by server.
	return p.observing(observer)
}

func (p *Peer) observing(observer Observer) error {
	for {
		// accept and pure the reader.
		r, err := p.conn.AcceptUniStream(context.Background())
		if err != nil {
			return err
		}

		// dispatch the reader and writer to the observer.
		go observer.Handle(p, r)
	}
}

func (p *Peer) Close() error {
	return p.conn.Close()
}
