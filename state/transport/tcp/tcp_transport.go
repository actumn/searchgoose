package tcp

import (
	"bytes"
	"encoding/gob"
	"github.com/actumn/searchgoose/state/transport"
	"io"
	"log"
	"net"
)

type RequestHandler func(conn net.Conn, req []byte) []byte

type Transport struct {
	LocalAddress    string
	LocalNodeId     string
	SeedHosts       []string
	RequestHandlers map[string]RequestHandler
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
			log.Fatalf("Fail to get response from %s; err: %v", address, err)
			return
		}

		data := DataFormatFromBytes(recvBuf[:n])
		callback(data.Content)
	}()
}

func NewTransport(address string, nodeId string, seedHosts []string) *Transport {
	return &Transport{
		LocalAddress: address,
		LocalNodeId:  nodeId,
		SeedHosts:    seedHosts,
	}
}

func (t *Transport) Register(action string, handler RequestHandler) {
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
					recvData := DataFormatFromBytes(recvBuf[:n])
					log.Printf("Receive REQ from %s\n", recvData.Source)
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

	// t.ConnectedNodes[recvNode.Id] = &conn

	callback(&Connection{
		conn: conn,
	})
}

// TODO :: Action type string으로 바꾸기

const (
	HANDSHAKE_REQ  = "handshake_req"
	HANDSHAKE_ACK  = "handshake_ack"
	HANDSHAKE_FAIL = "handshake_fail"
	PEERFIND_REQ   = "peerfind_req"
	PEERFIND_ACK   = "peerfind_ack"
	PEERFIND_FAIL  = "peerfind_fail"
)

type DataFormat struct {
	Source  string
	Dest    string
	Action  string
	Content []byte
}

func (d *DataFormat) ToBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(d); err != nil {
		log.Fatalln(err)
	}
	return buffer.Bytes()
}

func DataFormatFromBytes(b []byte) *DataFormat {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data DataFormat
	if err := decoder.Decode(&data); err != nil {
		log.Fatalln(err)
	}
	return &data
}
