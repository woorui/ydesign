package frame

// Tag is used for data router
type Tag uint32

type HandshakeFrame struct {
	AuthName    string
	AuthPayload string
}

type ConnectionFrame struct {
	Name            string
	ClientID        string
	ClientType      byte
	ObserveDataTags []Tag
}

type GoawayFrame struct {
	Message string
}

type HandshakeAckFrame struct{}

type Encoder interface {
	Encode() []byte
}

// Writer is the interface that wraps the WriteFrame method.

// Writer writes Frame from frm to the underlying data stream.
type Writer interface {
	WriteFrame(frm Encoder) error
}

func NewGoawayFrame(msg string) *GoawayFrame { return &GoawayFrame{} }

func (f *GoawayFrame) Encode() []byte { return []byte{} }

func NewHandshakeAckFrame() *HandshakeAckFrame { return &HandshakeAckFrame{} }

func (f *HandshakeAckFrame) Encode() []byte { return []byte{} }
