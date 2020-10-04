package tcp

import (
	"bytes"
	"encoding/gob"
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
	conn         net.Conn
	localAddress string
	destAddress  string
}

func (c *Connection) SendRequest(action string, content []byte, callback func(byte []byte)) {

	request := DataFormat{
		Source:  c.GetSourceAddress(),
		Dest:    c.GetDestAddress(),
		Action:  action,
		Content: content,
	}
	log.Printf("Send %s to %s\n", request.Action, request.Dest)

	_, err := c.conn.Write(request.ToBytes())
	if err != nil {
		log.Printf("Fail to send request; err:%v\n", err)
	}
	go func() {
		recvBuf := make([]byte, 4096)
		n, err := c.conn.Read(recvBuf)
		if err != nil {
			// logrus.Fatalf("Fail to get response from %s; err: %v", address, err)
			logrus.Fatalf("Fail to get response; err: %v", err)
			return
		}
		response := DataFormatFromBytes(recvBuf[:n])
		log.Printf("Receive %s from %s\n", response.Action, response.Source)
		callback(response.Content)
	}()
}

func (c *Connection) GetSourceAddress() string {
	return c.localAddress
}

func (c *Connection) GetDestAddress() string {
	return c.destAddress
}

type ReplyChannel struct {
	conn         net.Conn
	localAddress string
	destAddress  string
}

func (c *ReplyChannel) SendMessage(action string, content []byte) (n int, err error) {

	request := DataFormat{
		Source:  c.GetSourceAddress(),
		Dest:    c.GetDestAddress(),
		Action:  action,
		Content: content,
	}

	logrus.Infof("Send %s Reply to %s\n", action, request.Dest)
	return c.conn.Write(request.ToBytes())
}

func (c *ReplyChannel) GetSourceAddress() string {
	return c.localAddress
}

func (c *ReplyChannel) GetDestAddress() string {
	return c.destAddress
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
		logrus.Infof("Success of listening on %s", address)
		defer l.Close()

		for {
			conn, err := l.Accept()
			if err != nil {
				logrus.Infof("Fail to accept; err: %v", err)
				continue
			}
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
				for {
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
						action := recvData.Action
						data := recvData.Content

						t.RequestHandlers[action](&ReplyChannel{
							conn:         conn,
							localAddress: t.LocalAddress,
							destAddress:  recvData.Source,
						}, data)
					}
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
	logrus.Info("Success on connecting ", address)

	c := &Connection{
		conn:         conn,
		localAddress: t.LocalAddress,
		destAddress:  address,
	}
	callback(c)
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

// Data Format
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
