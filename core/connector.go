package core

import "io"

type ClientType string

type ObversedTag uint64

type Connection interface {
	io.Closer

	ClientType() ClientType
	ClientID() string
	ObversedTags() []ObversedTag
	SourceID() string
}

type Connector struct {
}
