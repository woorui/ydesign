package core

import (
	"context"

	"github.com/lucas-clemente/quic-go"
)

type Type uint8

// Frame 只和数据相关，只负责 buf 的编解码
//
// 1. 需要返回 frame 的 type。
// 2. 在 server 端，对于不同的 type，有不同的处理器。
// 3. Server 端处理器的职责包括，解析 frame，处理 frame 的对应逻辑，给 stream 返回处理结果。
// 4. frame 还有可能返回一些元数据，绑定到 server 之上。（不作为 hander，作为拦截器）
// 5. Client 端要求 frame 可以写入 stream。
// 6. Frame 还可以返回 frame。
type Frame interface {
	// Type 返回 frame 的 type。
	// 在 server 端，对于不同的 type，有不同的处理器。
	Type() Type

	// Encode 编码为 frame
	Encode() []byte

	// Decode 解码为 []byte
	Decode([]byte) error
}

// FrameHandler 可以注册到多路复用器，多路复用器根据不同的 frame type，找到相应的处理方式
type ServerFrameHandler interface {
	FrameType() Type

	// Handle 负责解析 frame，处理 frame，并且写回返回信息到 stream。
	// 不对外暴露具体的 Frame 的类型。
	// 如果 hande 返回了错误，表示是致命错误，需要退出 server。
	Handle(context.Context, Connector, quic.Connection, quic.Stream) error
}

type ClientFrameHandler interface {
	FrameType() Frame

	// Handle 负责解析 frame，处理 frame，并且写回返回信息到 stream。
	// 不对外暴露具体的 Frame 的类型。
	Handle(context.Context, quic.Connection, quic.Stream)
}
