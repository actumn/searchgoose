package tcp

import (
	"github.com/actumn/searchgoose/state/transport"
	"github.com/sirupsen/logrus"
	"io"
	"net"
)

type Transport struct {
	LocalAddress    string
	LocalNodeId     string
	SeedHosts       []string
	RequestHandlers map[string]transport.RequestHandler
}

type Connection struct {
	conn net.Conn
}

func (c *Connection) SendRequest(action string, req []byte, callback func(byte []byte)) {
	c.conn.Write(req)

	go func() {
		recvBuf := make([]byte, 4096)
		n, err := c.conn.Read(recvBuf)
		if err != nil {
			// logrus.Fatalf("Fail to get response from %s; err: %v", address, err)
			logrus.Fatalf("Fail to get response; err: %v", err)
			return
		}

		data := transport.DataFormatFromBytes(recvBuf[:n])
		callback(data.Content)
	}()
}

type ReplyChannel struct {
	conn net.Conn
}

func (c *ReplyChannel) SendMessage(b []byte) (n int, err error) {
	return c.conn.Write(b)
}

func NewTransport(address string, nodeId string, seedHosts []string) *Transport {
	return &Transport{
		LocalAddress:    address,
		LocalNodeId:     nodeId,
		SeedHosts:       seedHosts,
		RequestHandlers: make(map[string]transport.RequestHandler),
	}
}

func (t *Transport) Register(action string, handler transport.RequestHandler) {
	t.RequestHandlers[action] = handler
}

func (t *Transport) Start(address string) {
	go func() {
		l, err := net.Listen("tcp", address)
		// l, err := net.Listen("tcp", ":8180")
		if err != nil {
			logrus.Fatalf("Fail to bind address to %s; err: %v", address, err)
		}
		logrus.Info("Success of listening on %s", address)
		defer l.Close()

		for {
			conn, err := l.Accept()
			if err != nil {
				logrus.Info("Fail to accept; err: %v", err)
				continue
			}
			// Connection Handler
			go func(conn net.Conn) {
				recvBuf := make([]byte, 4096)
				n, err := conn.Read(recvBuf)
				if err != nil {
					if io.EOF == err {
						logrus.Info("connection is closed from client; %v", conn.RemoteAddr().String())
						return
					}
					logrus.Info("Fail to receive data; err: %v", err)
					return
				}
				if 0 < n {
					// Receive request data
					recvData := transport.DataFormatFromBytes(recvBuf[:n])
					action := recvData.Action
					data := recvData.Content

					// Send response data
					message := t.RequestHandlers[action](&ReplyChannel{
						conn: conn,
					}, data)
					conn.Write(message)
				}
			}(conn)
		}
	}()
}

func (t *Transport) OpenConnection(address string, callback func(conn transport.Connection)) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		logrus.Fatalf("Failed to connect to %s : %v", address, err)
	}
	logrus.Info("Success on connecting %s", address)

	callback(&Connection{
		conn: conn,
	})
}

func (t *Transport) GetLocalAddress() string {
	return t.LocalAddress
}

func (t *Transport) GetSeedHosts() []string {
	return t.SeedHosts
}

func (t *Transport) GetHandler(action string) transport.RequestHandler {
	return t.RequestHandlers[action]
}
