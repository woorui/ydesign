package core

import (
	"io"

	"github.com/woorui/ydesign/core/frame"
)

type FrameReadWriter struct {
	codec       frame.Codec
	packetCodec frame.PacketCodec
}

func NewFrameReadWriter(packetCodec frame.PacketCodec, codec frame.Codec) *FrameReadWriter {
	return &FrameReadWriter{
		packetCodec: packetCodec,
		codec:       codec,
	}
}

func (rw *FrameReadWriter) Readframe(r io.Reader) (frame.Frame, error) {
	ft, raw, err := rw.packetCodec.ReadPacket(r)
	if err != nil {
		return nil, err
	}

	f, err := frame.NewFrame(ft)
	if err != nil {
		return nil, err
	}

	err = rw.codec.Decode(raw, f)

	return f, err
}

func (rw *FrameReadWriter) WriteFrame(w io.Writer, f frame.Frame) error {
	data, err := rw.codec.Encode(f)
	if err != nil {
		return err
	}
	return rw.packetCodec.WritePacket(w, f.Type(), data)
}
