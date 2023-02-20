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
	"github.com/woorui/ydesign/core/frame"
	"golang.org/x/exp/slog"
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
	handler     Handler
	streamGroup sync.WaitGroup
	logger      *slog.Logger
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
	listener, err := quic.Listen(conn, s.TLSConfig, s.QuicConfig)
	if err != nil {
		return err
	}

	s.listener = listener

	for {
		// TODO: trackConn, make it don't accept if server is closing.
		conn, err := s.listener.Accept(ctx)
		if err != nil {
			return err
		}
		go s.handleConn(ctx, conn)
	}

	// TODO: idle conn gc.

}

func (s *Server) handleConn(ctx context.Context, conn quic.Connection) error {
	stream0, err := conn.AcceptStream(ctx)
	if err != nil {
		return err
	}

	controlStream := NewControlStream(stream0)

	signal, err := controlStream.AcceptSignaling(func(stream quic.Stream) (frame.Tag, []byte) { return frame.Tag(1), []byte{} })
	if err != nil {
		return nil
	}

	fmt.Println(signal)

	for {
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
