package core

import (
	"io"

	"github.com/lucas-clemente/quic-go"
)

type ClientType string

type ObversedTag uint64

type Connection interface {
	io.Closer

	Authenticated() bool

	ConnMetadata() ConnMetadata
}

type ConnMetadata interface {
	Name() string
	ClientType() ClientType
	ClientID() string
	ObversedTags() []ObversedTag
	SourceID() string
}

type Auth interface{}

// 这里是否只应该索引 connID，quic.Connection 的管理应该交给 server ？
type Connector interface {
	// Connector 需要优雅关闭
	// 在 Connector 关闭期间，新的链接不可以被添加到 Connector 并且所有 api 会被冻结
	// 如果 Connector 已经被关闭，返回 `ErrConnectorClosed`.
	io.Closer

	// Add 添加 Connection 到 Connector。
	// Add 根据 连接信息，元信息和认证信息构建 Connection
	// 并构建 Connection 添加到 Connector
	// 不允许重名，重名返回 DuplicateName
	// auth 失败返回 auth 失败错误，并返回构建好的连接
	Add(quic.Connection, quic.Stream, ConnMetadata, Auth) ConnResult

	GetByConnID(connID string) Connection

	GetByTagAndSourceID(tags []ObversedTag, sourceID string) []Connection
}

type ConnResult struct {
	Err           error
	DuplicateName bool
	Conn          Connection
}
