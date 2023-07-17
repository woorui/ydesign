package core

import (
	"context"
	"io"
	"net"
)

// A Listener accepts connections.
type Listener interface {
	// Close the server. All active connections will be closed.
	Close() error
	// Addr returns the local network addr that the server is listening on.
	Addr() net.Addr
	// Accept returns new connections. It should be called in a loop.
	Accept(context.Context) (Connection, error)
}

// A Connection is a connection between two peers.
type Connection interface {
	// LocalAddr returns the local address.
	LocalAddr() string
	// RemoteAddr returns the address of the peer.
	RemoteAddr() string
	// OpenStream opens a new bidirectional QUIC stream.
	OpenStream() (io.ReadWriteCloser, error)
	// AcceptStream returns the next stream opened by the peer, blocking until one is available.
	// If the connection was closed due to a timeout, the error satisfies the net.Error interface, and Timeout() will be true.
	AcceptStream(context.Context) (io.ReadWriteCloser, error)
	// OpenUniStream opens a new unbidirectional QUIC stream.
	OpenUniStream() (io.WriteCloser, error)
	// AcceptUniStream returns the next unidirectional stream opened by the peer, blocking until one is available.
	AcceptUniStream(context.Context) (io.ReadCloser, error)
	// CloseWithError closes the connection with an error.
	CloseWithError(string) error
}
