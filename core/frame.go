package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/yomorun/y3"
)

type Type uint8

type MockFrame struct {
	buf bytes.Buffer
}

// FrameType implements Frame
func (f *MockFrame) FrameType() Type {
	return 1
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

// 需求是 frame 可以注册到多路复用器，多路复用器根据不同的frame，找到相应的处理方式
// 1. 需要返回 frame 的 type。
// 2. 在 server 端，对于不同的 type，有不同的处理器。
// 3. Server 端处理器的职责包括，解析 frame，处理 frame 的对应逻辑，给 stream 返回处理结果。
// 4. Client 端要求 frame 可以写入 stream。
type Frame interface {
	// Type 返回 frame 的 type。
	// 在 server 端，对于不同的 type，有不同的处理器。
	Type() Type

	// Handle 负责解析 frame，处理 frame，并且写回返回信息到 stream
	Handle(context.Context, io.WriteCloser)

	// Encode 解析到 byte 数组
	Encode() ([]byte, error)
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
