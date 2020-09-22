package tcp

import (
	"github.com/actumn/searchgoose/state/transport"
	"io"
	"log"
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

func (c *Connection) SendRequest(req []byte, callback func(byte []byte)) {

	c.conn.Write(req)

	recvBuf := make([]byte, 4096)

	go func() {
		n, err := c.conn.Read(recvBuf)

		if err != nil {
			// log.Fatalf("Fail to get response from %s; err: %v", address, err)
			log.Fatalf("Fail to get response; err: %v", err)
			return
		}

		data := transport.DataFormatFromBytes(recvBuf[:n])
		callback(data.Content)
	}()
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
			log.Fatalf("Fail to bind address to %s; err: %v", address, err)
		}
		log.Printf("Success of listening on %s\n", address)
		defer l.Close()

		for {
			conn, err := l.Accept()
			if err != nil {
				log.Printf("Fail to accept; err: %v", err)
				continue
			}
			// Connection Handler
			go func(conn net.Conn) {
				recvBuf := make([]byte, 4096)
				n, err := conn.Read(recvBuf)
				if err != nil {
					if io.EOF == err {
						log.Printf("connection is closed from client; %v", conn.RemoteAddr().String())
						return
					}
					log.Printf("Fail to receive data; err: %v", err)
					return
				}
				if 0 < n {
					// Receive request data
					recvData := transport.DataFormatFromBytes(recvBuf[:n])
					action := recvData.Action
					data := recvData.Content

					// Send response data
					message := t.RequestHandlers[action](conn, data)
					conn.Write(message)
				}
			}(conn)
		}
	}()
}

func (t *Transport) OpenConnection(address string, callback func(conn transport.Connection)) {

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("Failed to connect to %s : %v", address, err)
	}
	log.Printf("Success on connecting %s\n", address)

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
