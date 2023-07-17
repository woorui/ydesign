package frame

import (
	"fmt"
	"io"
)

// Frame is the minimum unit required for Yomo to run.
// Yomo transmits various instructions and data through the frame, which can be transmitted on the IO stream.
type Frame interface {
	// Type returns the type of frame.
	Type() Type
}

// Type defined The type of frame.
type Type byte

// PacketCodec reads packet from the io.Reader and writes packet to the io.Writer.
// It returns frameType, the data of the packet and an error if read failed.
type PacketCodec interface {
	// ReadPacket reads raw frame from Reader.
	ReadPacket(io.Reader) (Type, []byte, error)
	// WritePacket writes raw frame from Writer.
	WritePacket(io.Writer, Type, []byte) error
}

// Codec encodes and decodes byte array (the raw frame data) to frame.
type Codec interface {
	// Decode decodes byte array to frame.
	Decode([]byte, Frame) error
	// Encode encodes frame to byte array.
	Encode(Frame) ([]byte, error)
}

// AuthenticationFrame is used to authenticate the client,
// Once the connection is established, the client immediately sends information to the server,
// server gets the way to authenticate according to AuthName and use AuthPayload to do a authentication.
// AuthenticationFrame is transmit on ControlStream.
//
// Reading the `auth.Authentication` interface will help you understand how AuthName and AuthPayload work.
type AuthenticationFrame struct {
	// AuthName.
	AuthName string
	// AuthPayload.
	AuthPayload string
}

// Type returns the type of AuthenticationFrame.
func (f *AuthenticationFrame) Type() Type { return TypeAuthenticationFrame }

// AuthenticationAckFrame is used to confirm that the client is authorized to access the requested DataStream from
// ControlStream, AuthenticationAckFrame is transmit on ControlStream.
// If the client-side receives this frame, it indicates that authentication was successful.
type AuthenticationAckFrame struct {
	// ID is the ID of the ControlStream. It is assigned by the server.
	ID string
}

// Type returns the type of AuthenticationAckFrame.
func (f *AuthenticationAckFrame) Type() Type { return TypeAuthenticationAckFrame }

// ObserveFrame is used to open a new peer stream and make connection observe a peer stream.
// Each peer stream has a tag that identifies which connection can observe it.
type ObserveFrame struct {
	// Tag is used to identify the controlStream observer associated with a particular peer stream.
	Tag string
}

// Type returns the type of ObserveFrame.
func (f *ObserveFrame) Type() Type { return TypeObserveFrame }

// OpenStreamFrame is used to open a new peer stream and tells which connection can observe it.
// Each peer stream opens a stream that has a identifier.
type OpenStreamFrame struct {
	// ID is the identifier that be used to identify the stream be opened by peer.
	ID string
	// Tag is used to identify the controlStream observer associated with a particular peer stream.
	Tag string
	// Metadata is the metadata of stream be created.
	Metadata []byte
}

// Type returns the type of OpenStreamFrame.
func (f *OpenStreamFrame) Type() Type { return TypeOpenStreamFrame }

// RejectedFrame is is used to reject a ControlStream reqeust.
type RejectedFrame struct {
	// Code is the code rejected.
	Code uint64
	// Message contains the reason why the reqeust be rejected.
	Message string
}

// Type returns the type of RejectedFrame.
func (f *RejectedFrame) Type() Type { return TypeRejectedFrame }

const (
	TypeAuthenticationFrame    Type = 0x03 // TypeAuthenticationFrame is the type of AuthenticationFrame.
	TypeAuthenticationAckFrame Type = 0x11 // TypeAuthenticationAckFrame is the type of AuthenticationAckFrame.
	TypeRejectedFrame          Type = 0x39 // TypeRejectedFrame is the type of RejectedFrame.
	TypeObserveFrame           Type = 0x2F // TypeObserveFrame is the type of ObserveFrame.
	TypeOpenStreamFrame        Type = 0x30 // TypeOpenStreamFrame is the type of OpenStreamFrame
)

var frameTypeStringMap = map[Type]string{
	TypeAuthenticationFrame:    "AuthenticationFrame",
	TypeAuthenticationAckFrame: "AuthenticationAckFrame",
	TypeRejectedFrame:          "RejectedFrame",
	TypeObserveFrame:           "ObserveFrame",
	TypeOpenStreamFrame:        "OpenStreamFrame",
}

// String returns a human-readable string which represents the frame type.
// The string can be used for debugging or logging purposes.
func (f Type) String() string {
	frameString, ok := frameTypeStringMap[f]
	if ok {
		return frameString
	}
	return "UnkonwnFrame"
}

var frameTypeNewFuncMap = map[Type]func() Frame{
	TypeAuthenticationFrame:    func() Frame { return new(AuthenticationFrame) },
	TypeAuthenticationAckFrame: func() Frame { return new(AuthenticationAckFrame) },
	TypeObserveFrame:           func() Frame { return new(ObserveFrame) },
	TypeOpenStreamFrame:        func() Frame { return new(OpenStreamFrame) },
	TypeRejectedFrame:          func() Frame { return new(RejectedFrame) },
}

// NewFrame creates a new frame from Type.
func NewFrame(t Type) (Frame, error) {
	newFunc, ok := frameTypeNewFuncMap[t]
	if ok {
		return newFunc(), nil
	}
	return nil, fmt.Errorf("frame: cannot new a frame from %d", int64(t))
}
