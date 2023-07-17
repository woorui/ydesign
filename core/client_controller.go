package core

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/quic-go/quic-go"
	"github.com/woorui/ydesign/core/auth"
	"github.com/woorui/ydesign/core/frame"
	"golang.org/x/exp/slog"
)

// clientController is the struct that defines the methods for client-side controler.
type clientController struct {
	ctx context.Context

	conn    quic.Connection
	stream0 quic.Stream

	// encode and decode the frame.
	frw *FrameReadWriter

	// id is the connID of the client, It is assigned by server.
	id string

	logger *slog.Logger
}

// ClientController is the interface that defines the methods for client-side control controller.
type ClientController interface {
	// Authenticate sends the provided credential to the server's control stream to authenticate the client.
	// There will return `ErrAuthenticateFailed` if authenticate failed.
	Authenticate(*auth.Credential) error

	// RequestObserve requests server to observe a tag.
	RequestObserve(tag string) error

	// CloseWithError closes the client-side control stream.
	CloseWithError(string) error
}

// ClientControllerOpener opens a client-side controller.
type ClientControllerOpener interface {
	// Open is the open function.
	Open(ctx context.Context, addr string) (ClientController, error)
}

// ErrAuthenticateFailed be returned when client control stream authenticate failed.
type ErrAuthenticateFailed struct {
	ReasonFromeServer string
}

// Error returns a string that represents the ErrAuthenticateFailed error for the implementation of the error interface.
func (e ErrAuthenticateFailed) Error() string { return e.ReasonFromeServer }

type ClientControlStreamOpener struct {
	tlsConfig  *tls.Config
	quicConfig *quic.Config

	logger *slog.Logger
}

func (opener *ClientControlStreamOpener) Open(ctx context.Context, addr string) (ClientController, error) {
	conn, err := quic.DialAddr(ctx, addr, opener.tlsConfig, opener.quicConfig)
	if err != nil {
		return nil, err
	}
	stream0, err := conn.OpenStream()
	if err != nil {
		return nil, err
	}
	// TODO:
	fmt.Println(stream0)

	return nil, nil
}

// RequestObserve requests server to observe a tag.
func (cs *clientController) RequestObserve(tag string) error {
	f := &frame.ObserveFrame{
		Tag: tag,
	}
	return cs.frw.WriteFrame(cs.stream0, f)
}

// ID returns the ID of the connection which is assigned by server.
func (cs *clientController) ID() string {
	return cs.id
}

// OpenUniStream opens a Writer.
func (cs *clientController) OpenUniStream() (io.WriteCloser, error) {
	return cs.conn.OpenUniStream()
}

// AcceptUniStream accepts a Reader.
func (cs *clientController) AcceptUniStream(ctx context.Context) (io.Reader, error) {
	return cs.conn.AcceptUniStream(ctx)
}

// NewClientController returns ClientController from quic Connection and the first stream form the Connection.
func NewClientController(
	ctx context.Context,
	conn quic.Connection, stream0 quic.Stream,
	frw *FrameReadWriter, logger *slog.Logger) *clientController {

	controlStream := &clientController{
		ctx:     ctx,
		conn:    conn,
		stream0: stream0,
		frw:     frw,
		id:      "", // there is empty id if not being authenticated.
		logger:  logger,
	}

	return controlStream
}

func (cs *clientController) readFrameLoop() {
	for {
		f, err := cs.frw.Readframe(cs.stream0)
		if err != nil {
			// TODO: new code!
			cs.conn.CloseWithError(serverCloseCode, err.Error())
			return
		}
		switch ff := f.(type) {
		case *frame.RejectedFrame:
			// TODO: adapt to quic code.
			cs.conn.CloseWithError(quic.ApplicationErrorCode(ff.Code), ff.Message)
			return
		default:
			cs.logger.Debug("control stream read unexcepted frame", "frame_type", f.Type().String())
		}
	}
}

// Authenticate sends the provided credential to the server's control stream to authenticate the client.
// There will return `ErrAuthenticateFailed` if authenticate failed.
func (cs *clientController) Authenticate(cred auth.Credential) error {
	af := &frame.AuthenticationFrame{
		AuthName:    cred.Name(),
		AuthPayload: cred.Payload(),
	}
	if err := cs.frw.WriteFrame(cs.stream0, af); err != nil {
		return err
	}
	received, err := cs.frw.Readframe(cs.stream0)
	if err != nil {
		if qerr := new(quic.ApplicationError); errors.As(err, &qerr) && strings.HasPrefix(qerr.ErrorMessage, "authentication failed") {
			return &ErrAuthenticateFailed{qerr.ErrorMessage}
		}
		return err
	}
	f, ok := received.(*frame.AuthenticationAckFrame)
	if !ok {
		return fmt.Errorf(
			"yomo: read unexpected frame during waiting authentication resp, frame read: %s",
			received.Type().String(),
		)
	}
	cs.id = f.ID

	// create a goroutinue to continuous read frame from server.
	go cs.readFrameLoop()

	return nil
}

// Observe tells server that the client wants to observe specified tag.
func (cs *clientController) Observe(tag string) error {
	f := &frame.ObserveFrame{
		Tag: tag,
	}
	return cs.frw.WriteFrame(cs.stream0, f)
}

// CloseWithError closes the client-side control stream.
func (cs *clientController) CloseWithError(errString string) error {
	cs.stream0.Close()
	return cs.conn.CloseWithError(serverCloseCode, errString)
}
