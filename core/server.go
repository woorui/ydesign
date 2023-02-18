package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"runtime"
	"sync"

	"github.com/quic-go/quic-go"
)

type Handler interface {
	ServeYomo(context.Context, quic.Connection, quic.Stream) error
}

type Server struct {
	// TLSConfig provides a TLS configuration for use by server. It must be
	// set for ListenAndServe and Serve methods.
	TLSConfig *tls.Config

	// QuicConfig provides the parameters for QUIC connection created with
	// Serve. If nil, it uses reasonable default values.
	QuicConfig *quic.Config

	Addr        string
	listener    quic.Listener
	onShutdown  []func()
	handler     Handler
	streamGroup sync.WaitGroup
}

// DefaultListenAddr is the default address to listen.
const DefaultListenAddr = "0.0.0.0:9000"

func NewServer(handler Handler, tlsConfig *tls.Config, quicConfig *quic.Config) *Server {
	return &Server{
		handler:    handler,
		TLSConfig:  tlsConfig,
		QuicConfig: quicConfig,
	}
}

func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	// TODO: add some info to ctx
	if addr == "" {
		addr = DefaultListenAddr
	}
	s.Addr = addr

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	return s.Serve(ctx, conn)
}

func (s *Server) Serve(ctx context.Context, conn net.PacketConn) error {
	l, err := quic.Listen(conn, s.TLSConfig, s.QuicConfig)
	if err != nil {
		return err
	}

	s.listener = l

	for {
		// TODO: trackConn, make it don't accept if server is closing.
		conn, err := l.Accept(ctx)
		if err != nil {
			return err
		}
		go s.handleConn(ctx, conn)
	}

	// TODO: idle conn gc.

}

func (s *Server) handleConn(ctx context.Context, conn quic.Connection) error {
	// TODO: check whether the conn is idle, if so, clise it.
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fmt.Println("........")
			stream, err := conn.AcceptStream(ctx)
			if err != nil {
				fmt.Println(err)
				return err
			}
			fmt.Println("streaming...")
			s.streamGroup.Add(1)
			go func() {
				defer func() {
					if err := recover(); err != nil {
						const size = 64 << 10
						buf := make([]byte, size)
						buf = buf[:runtime.Stack(buf, false)]
						log.Printf("yomo: panic serving %v: %v\n%s", conn.RemoteAddr(), err, buf)
					}
					s.streamGroup.Done()
				}()
				fmt.Println("serving...")
				s.handler.ServeYomo(ctx, conn, stream)
			}()
		}
	}
}

func (s *Server) Close() error {
	s.streamGroup.Wait()

	for _, fn := range s.onShutdown {
		go fn()
	}

	// TODO: trackConn, close every conn

	var err error
	if s.listener != nil {
		err = s.listener.Close()
	}

	return err
}
