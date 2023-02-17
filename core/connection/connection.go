package connection

import (
	"io"

	"github.com/woorui/ydesign/core/frame"
	"github.com/woorui/ydesign/core/metadata"
	"golang.org/x/exp/slog"
)

// Info holds connection informations.
type Info interface {
	// Name returns the name of the connection, which is set by clients.
	Name() string
	// ClientID connection client ID
	ClientID() string
	// ClientType returns the type of the client (Source | SFN | UpstreamZipper)
	ClientType() byte
	// Metadata returns the extra info of the application
	Metadata() metadata.Metadata
}

// Connection wraps the specific io connections (typically quic.Connection) to transfer y3 frames
type Connection interface {
	io.Closer

	Info

	// Write writes frame to underlying stream.
	// Write should goroutine-safely send y3 frames to peer side.
	frame.Writer

	// ObserveDataTags observed data tags
	ObserveDataTags() []frame.Tag
}

func New(
	name string,
	clientID string,
	clientType byte,
	metadata metadata.Metadata,
	stream io.ReadWriteCloser,
	observed []frame.Tag,
	logger *slog.Logger,
) Connection {
	return nil
}
