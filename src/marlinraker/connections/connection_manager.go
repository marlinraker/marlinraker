package connections

import (
	"github.com/gorilla/websocket"
	"github.com/samber/lo"
	"marlinraker/src/util"
	"sync"
	"sync/atomic"
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
	connections     = util.NewThreadSafe(make([]Connection, 0))
	nextWebsocketId atomic.Int32
)

func GetConnections() []Connection {
	return connections.Load()
}

func RegisterConnection(socket *websocket.Conn) *Connection {
	id := nextWebsocketId.Add(1)
	connection := Connection{
		socket: socket,
		mutex:  &sync.Mutex{},
		Id:     int(id),
	}
	connections.Do(func(connections []Connection) []Connection {
		return append(connections, connection)
	})
	return &connection
}

func UnregisterConnection(connection *Connection) {
	connections.Do(func(connections []Connection) []Connection {
		return lo.Filter(connections, func(_connection Connection, _ int) bool {
			return _connection.Id != connection.Id
		})
	})
}

func TerminateAllConnections() {
	connections.Do(func(connections []Connection) []Connection {
		for _, connection := range connections {
			if err := connection.socket.Close(); err != nil {
				util.LogError(err)
			}
		}
		return make([]Connection, 0)
	})
}
