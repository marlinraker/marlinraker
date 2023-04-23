package connections

import (
	"github.com/gorilla/websocket"
	"github.com/samber/lo"
	"sync"
)

type Connection struct {
	socket     *websocket.Conn
	mutex      *sync.Mutex
	Identified bool
	Id         int
	ClientName string
	Version    string
	ClientType string
	Url        string
}

func (connection *Connection) WriteText(bytes []byte) error {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()
	return connection.socket.WriteMessage(websocket.TextMessage, bytes)
}

func (connection *Connection) WriteJson(v any) error {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()
	return connection.socket.WriteJSON(v)
}

var (
	connections      = make([]*Connection, 0)
	connectionsMutex = &sync.RWMutex{}
	nextWebsocketId  = 0
)

func GetConnections() []*Connection {
	connectionsMutex.RLock()
	defer connectionsMutex.RUnlock()
	return connections
}

func RegisterConnection(socket *websocket.Conn) *Connection {
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()

	id := nextWebsocketId
	nextWebsocketId++
	connection := &Connection{
		socket: socket,
		mutex:  &sync.Mutex{},
		Id:     id,
	}
	connections = append(connections, connection)
	return connection
}

func UnregisterConnection(connection *Connection) {
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()
	connections = lo.Filter(connections, func(_connection *Connection, _ int) bool { return _connection != connection })
}
