package connections

import (
	"fmt"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Connection struct {
	socket     *websocket.Conn
	mutex      sync.Mutex
	infoMutex  sync.RWMutex
	Id         int
	identified bool
	clientInfo ClientInfo
}

type ClientInfo struct {
	ClientName string
	Version    string
	ClientType string
	Url        string
}

func (connection *Connection) WriteText(bytes []byte) error {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()
	err := connection.socket.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	return nil
}

func (connection *Connection) WriteJson(v any) error {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()
	err := connection.socket.WriteJSON(v)
	if err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}
	return nil
}

// Identify stores the client information of this connection. It returns false
// if the connection has already been identified.
func (connection *Connection) Identify(info ClientInfo) bool {
	connection.infoMutex.Lock()
	defer connection.infoMutex.Unlock()
	if connection.identified {
		return false
	}
	connection.identified, connection.clientInfo = true, info
	return true
}

// ClientInfo returns the information the client sent with server.connection.identify.
func (connection *Connection) ClientInfo() (ClientInfo, bool) {
	connection.infoMutex.RLock()
	defer connection.infoMutex.RUnlock()
	return connection.clientInfo, connection.identified
}

var (
	connectionsMutex sync.RWMutex
	connections      []*Connection
	nextWebsocketId  atomic.Int32
)

// GetConnections returns a snapshot of all currently registered connections.
func GetConnections() []*Connection {
	connectionsMutex.RLock()
	defer connectionsMutex.RUnlock()
	return slices.Clone(connections)
}

// FindConnection returns the registered connection with the given id.
func FindConnection(id int) (*Connection, bool) {
	connectionsMutex.RLock()
	defer connectionsMutex.RUnlock()
	idx := slices.IndexFunc(connections, func(connection *Connection) bool {
		return connection.Id == id
	})
	if idx == -1 {
		return nil, false
	}
	return connections[idx], true
}

func RegisterConnection(socket *websocket.Conn) *Connection {
	connection := &Connection{
		socket: socket,
		Id:     int(nextWebsocketId.Add(1)),
	}
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()
	connections = append(connections, connection)
	return connection
}

func UnregisterConnection(connection *Connection) {
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()
	connections = slices.DeleteFunc(connections, func(other *Connection) bool {
		return other == connection
	})
}

func TerminateAllConnections() {
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()
	for _, connection := range connections {
		if err := connection.socket.Close(); err != nil {
			log.Errorf("Failed to close socket %s: %v", connection.socket.RemoteAddr(), err)
		}
	}
	connections = nil
}
