package internal

import (
	"bytes"
	"fmt"
	"io"

	"github.com/yomorun/y3"
)

type Type uint8

type MockFrame struct {
	buf bytes.Buffer
}

// FrameType implements Frame
func (f *MockFrame) FrameType() FrameType {
	return "mock_type"
}

// ReadFrom implements Frame
func (f *MockFrame) ReadFrom(stream io.Reader) (int64, error) {
	fmt.Println("read")
	return io.Copy(&f.buf, stream)
}

// WriteTo implements Frame
func (f *MockFrame) WriteTo(stream io.Writer) (int64, error) {
	fmt.Println("write")
	f.buf.WriteString("wwwwww")
	return io.Copy(stream, &f.buf)
}

type RejectedFrame struct {
	message string
}

// NewRejectedFrame creates a new RejectedFrame with a given TagID of user's data
func NewRejectedFrame(msg string) *RejectedFrame {
	return &RejectedFrame{message: msg}
}

const TagOfRejectedMessage Type = 0x02

// Type gets the type of Frame.
func (f *RejectedFrame) Type() Type { return 1 }

// Encode to Y3 encoded bytes
func (f *RejectedFrame) Encode() []byte {
	rejected := y3.NewNodePacketEncoder(byte(f.Type()))
	// message
	msgBlock := y3.NewPrimitivePacketEncoder(byte(TagOfRejectedMessage))
	msgBlock.SetStringValue(f.message)

	rejected.AddPrimitivePacket(msgBlock)

	return rejected.Encode()
}

// func (f *RejectedFrame) ReadFrom(stream io.Reader) (int64, error) {}

// func (f *RejectedFrame) WriteTo(stream io.Writer) (int64, error) {}

// Message rejected message
func (f *RejectedFrame) Message() string {
	return f.message
}

// DecodeToRejectedFrame decodes Y3 encoded bytes to RejectedFrame
func DecodeToRejectedFrame(buf []byte) (*RejectedFrame, error) {
	node := y3.NodePacket{}
	_, err := y3.DecodeToNodePacket(buf, &node)
	if err != nil {
		return nil, err
	}
	rejected := &RejectedFrame{}
	// message
	if msgBlock, ok := node.PrimitivePackets[byte(TagOfRejectedMessage)]; ok {
		msg, err := msgBlock.ToUTF8String()
		if err != nil {
			return nil, err
		}
		rejected.message = msg
	}
	return rejected, nil
}
