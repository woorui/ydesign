package core

import (
	"errors"
	"fmt"

	"github.com/quic-go/quic-go"
	"github.com/woorui/ydesign/core/connection"
	"github.com/woorui/ydesign/core/frame"
	"github.com/woorui/ydesign/core/metadata"
	"github.com/woorui/ydesign/core/router"
)

type internalServer interface {
	getQuicConnection() quic.Connection
	getControlStream() quic.Stream
	getAuthenticate(*frame.HandshakeFrame) bool
	getMedataBuilder() metadata.Builder
	getRouter() router.Router
	getConnector() Connector
}

type ControlStream struct {
	stream quic.Stream
}

func NewControlStream(stream quic.Stream) *ControlStream {
	return &ControlStream{stream}
}

type FrameReadFunc func(stream quic.Stream) (frame.Tag, []byte, error)

func (cs *ControlStream) AcceptSignaling(fn FrameReadFunc) (Signaling, error) {
	tag, buf, err := fn(cs.stream)
	if err != nil {
		return nil, err
	}

	switch tag {
	case frame.HandshakeFrameTag:
		f, err := frame.DecodeToHandshakeFrame(buf)
		if err != nil {
			return nil, err
		}
		return SignalingHandshake(f), nil
	}

	return nil, errors.New("unexpect frame read")
}

type Signaling func(srv internalServer, c *Context) error

func SignalingHandshake(f *frame.HandshakeFrame) Signaling {
	return func(srv internalServer, c *Context) error {
		if ok := srv.getAuthenticate(f); ok {
			c.Logger.Debug("Authentication succeeded")
			return nil
		}
		c.Logger.Debug("Authentication failed")
		errString := fmt.Sprintf("handshake authentication failed, client credential name is %s", f.AuthName)

		goaway := frame.NewGoawayFrame(errString)

		if _, err := srv.getControlStream().Write(goaway.Encode()); err != nil {
			return err
		}
		if err := srv.getQuicConnection().CloseWithError(0, errString); err != nil {
			return err
		}

		ack := frame.NewHandshakeAckFrame()

		_, err := srv.getControlStream().Write(ack.Encode())
		return err
	}
}

func SignalingConnection(f *frame.ConnectionFrame) Signaling {
	return func(srv internalServer, c *Context) error {
		stream, err := srv.getQuicConnection().OpenStream()
		if err != nil {
			return err
		}

		c.Set(ConnectionInfoKey, f)

		var (
			connID          = f.ClientID
			name            = f.Name
			observeDataTags = f.ObserveDataTags
			clientType      = ClientType(f.ClientType)
		)

		metadata, err := srv.getMedataBuilder().Build(f)
		if err != nil {
			return err
		}

		conn := connection.New(name, connID, byte(clientType), metadata, stream, observeDataTags, c.Logger)

		// stream function route
		if clientType == ClientTypeStreamFunction {
			route := srv.getRouter().Route(metadata)

			if route == nil {
				return errors.New("connection route is nil")
			}
			// There should be Set api.
			if err := route.Add(connID, name, observeDataTags); err != nil {
				return err
			}
		}

		srv.getConnector().Add(connID, conn)

		c.Logger.Info("Connected!")

		return nil
	}
}
