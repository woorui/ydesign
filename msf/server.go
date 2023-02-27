package msf

import "github.com/quic-go/quic-go"

type Server struct {
	conn          quic.Connection
	controlStream quic.Stream
}
