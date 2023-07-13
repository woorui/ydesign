package stateful

import (
	"context"
	"io"

	"golang.org/x/exp/slog"
)

// Client represents a peer in a network that can open writers and consume other writers,
// and handle them in an consumer.
type Client struct {
	conn           UniStreamPeerConnection
	logger         *slog.Logger
	fillWriterFunc func(string, string, io.Writer) error
	idGenerator    func() string
}

// NewClient returns a new Client from a connection.
func NewClient(
	conn UniStreamPeerConnection,
	logger *slog.Logger,
	fillWriterFunc func(string, string, io.Writer) error,
	idGenerator func() string,
) *Peer {
	peer := &Peer{
		conn:           conn,
		logger:         logger,
		fillWriterFunc: fillWriterFunc,
		idGenerator:    idGenerator,
	}

	return peer
}

// Open opens a writer with the given tag, which other peers can observe.
// The returned writer can be used to write to the stream associated with the given tag.
func (p *Peer) Open(tag string) (io.WriteCloser, error) {
	w, err := p.conn.OpenUniStream()
	if err != nil {
		return nil, err
	}

	p.logger.Debug("peer opens a writer", "tag", tag)

	id := p.idGenerator()

	return w, p.fillWriterFunc(id, tag, w)
}

// Observe observes tagged streams and handles them in an observer.
// The observer is responsible for handling the tagged streams and writing to a new peer stream.
func (p *Peer) Observe(tag string, observer Observer) error {
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
