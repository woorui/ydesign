package core

import (
	"github.com/woorui/ydesign/core/connection"
	"github.com/woorui/ydesign/core/frame"
)

type ObversedTag uint64

// Connector is a interface to manage the connections and applications.
type Connector interface {
	// Add a connection.
	Add(connID string, conn connection.Connection)
	// Remove a connection.
	Remove(connID string)
	// Get a connection by connection id.
	Get(connID string) connection.Connection
	// GetSnapshot gets the snapshot of all connections.
	GetSnapshot() map[string]string
	// GetSourceConns gets the connections by source observe tag.
	GetSourceConns(sourceID string, tag frame.Tag) []connection.Connection
	// Clean the connector.
	Clean()
}
