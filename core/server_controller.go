package core

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/quic-go/quic-go"
	"github.com/woorui/ydesign/core/frame"
	"github.com/woorui/ydesign/core/metadata"
	"golang.org/x/exp/slog"
)

// ServerController defines the struct of server-side control stream.
type ServerController struct {
	id      string
	ctx     context.Context
	server  *Server
	conn    quic.Connection
	stream0 quic.Stream
	frw     *FrameReadWriter

	observeChan chan string

	logger *slog.Logger
}

const serverCloseCode = quic.ApplicationErrorCode(0xDF)

// NewServerControlStream returns ServerControlStream from the first stream of this Connection.
func NewServerController(
	ctx context.Context,
	conn quic.Connection, stream0 quic.Stream,
	frw *FrameReadWriter, logger *slog.Logger, idGenerator func() string) *ServerController {
	controller := &ServerController{
		id:      idGenerator(),
		ctx:     ctx,
		stream0: stream0,
		frw:     frw,
		logger:  logger,
	}

	return controller
}

func (ss *ServerController) readFrameLoop() {
	defer func() {
		close(ss.observeChan)
	}()
	for {
		f, err := ss.frw.Readframe(ss.stream0)
		if err != nil {
			return
		}
		switch ff := f.(type) {
		case *frame.ObserveFrame:
			ss.observeChan <- ff.Tag
		default:
			ss.logger.Debug("control stream read unexpected frame", "frame_type", f.Type().String())
		}
	}
}

func (ss *ServerController) ID() string {
	return ss.id
}

func (ss *ServerController) AcceptUniStream(ctx context.Context) (io.Reader, error) {
	return ss.conn.AcceptUniStream(ctx)
}

func (ss *ServerController) OpenUniStream() (io.WriteCloser, error) {
	return ss.conn.OpenUniStream()
}

// CloseWithError closes the server-side control stream.
func (ss *ServerController) CloseWithError(errString string) error {
	return ss.conn.CloseWithError(serverCloseCode, errString)
}

// VerifyAuthenticationFunc is used by server control stream to verify authentication.
type VerifyAuthenticationFunc func(*frame.AuthenticationFrame) (metadata.MD, bool, error)

// VerifyAuthentication verify the Authentication from client side.
func (ss *ServerController) VerifyAuthentication(verifyFunc VerifyAuthenticationFunc) (metadata.MD, error) {
	first, err := ss.frw.Readframe(ss.stream0)
	if err != nil {
		return nil, err
	}

	received, ok := first.(*frame.AuthenticationFrame)
	if !ok {
		errString := fmt.Sprintf("authentication failed: read unexcepted frame, frame read: %s", received.Type().String())
		ss.rejectWithCloseConn(224, errString)
		return nil, errors.New(errString)
	}

	md, ok, err := verifyFunc(received)
	if err != nil {
		return md, err
	}

	// authentication failed.
	if !ok {
		errString := fmt.Sprintf("authentication failed: client credential name is %s", received.AuthName)
		ss.rejectWithCloseConn(223, errString)
		return md, errors.New(errString)
	}

	// authentication successful.
	ack := &frame.AuthenticationAckFrame{
		ID: ss.id,
	}
	if err := ss.frw.WriteFrame(ss.stream0, ack); err != nil {
		return md, err
	}

	// create a goroutinue to continuous read frame after verify authentication successful.
	go ss.readFrameLoop()

	return md, nil
}

func (ss *ServerController) rejectWithCloseConn(code int64, msg string) {
	rejected := &frame.RejectedFrame{
		Code:    224,
		Message: msg,
	}

	err := ss.frw.WriteFrame(ss.stream0, rejected)
	if err != nil {
		ss.logger.Debug("server write rejected frame failed", "err", err)
	}

	if err = ss.CloseWithError(msg); err != nil {
		ss.logger.Debug("server colse rejected conn connection failed", "err", err)
	}
}
