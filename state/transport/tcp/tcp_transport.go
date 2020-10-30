package tcp

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
)

var (
	requestIdGenerator uint64
)

type Transport struct {
	LocalAddress    string
	LocalNodeId     string
	SeedHosts       []string
	RequestHandlers map[string]transport.RequestHandler
}

type Connection struct {
	conn             net.Conn
	localAddress     string
	destAddress      string
	err              string
	responseHandlers map[uint64]func(byte []byte)
}

func (c *Connection) SendRequest(action string, content []byte, callback func(byte []byte)) {
	atomic.AddUint64(&requestIdGenerator, 1)

	request := DataFormat{
		Id:      requestIdGenerator,
		Source:  c.GetSourceAddress(),
		Dest:    c.GetDestAddress(),
		Action:  action,
		Content: content,
	}
	c.responseHandlers[request.Id] = callback
	logrus.Infof("Send %s to %s\n", request.Action, request.Dest)

	bytesBuf := request.toBytes()
	lengthBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthBuf, uint32(len(bytesBuf)))
	if _, err := c.conn.Write(lengthBuf); err != nil {
		logrus.Errorf("Failed to send msg length; err: %v", err)
	}
	if _, err := c.conn.Write(bytesBuf); err != nil {
		logrus.Errorf("Fail to send request; err:%v\n", err)
	}
	go func() {
		lengthBuf := make([]byte, 4)
		if _, err := c.conn.Read(lengthBuf); err != nil {
			logrus.Fatalf("Failed to read msg length; err: %v", err)
		}
		msgLength := binary.LittleEndian.Uint32(lengthBuf)
		recvBuf := make([]byte, int(msgLength))
		if _, err := io.ReadFull(c.conn, recvBuf); err != nil {
			logrus.Fatalf("Fail to get response; err: %v", err)
		}
		response := dataFormatFromBytes(recvBuf)
		logrus.Infof("Receive %s from %s\n", response.Action, response.Source)
		if strings.Contains(response.Action, "_FAIL") {
			logrus.Warnf("%s", string(response.Content))
		} else {
			handler := c.responseHandlers[response.Id]
			delete(c.responseHandlers, response.Id)
			handler(response.Content)
		}
	}()
}

func (c *Connection) GetSourceAddress() string {
	return c.localAddress
}

func (c *Connection) GetDestAddress() string {
	return c.destAddress
}

func (c *Connection) GetMessage() string {
	return c.err
}

type ReplyChannel struct {
	requestId    uint64
	conn         net.Conn
	localAddress string
	destAddress  string
}

func (c *ReplyChannel) SendMessage(action string, content []byte) (n int, err error) {

	request := DataFormat{
		Id:      c.requestId,
		Source:  c.GetSourceAddress(),
		Dest:    c.GetDestAddress(),
		Action:  action,
		Content: content,
	}

	logrus.Infof("Send %s Reply to %s\n", action, request.Dest)

	bytesBuf := request.toBytes()
	lengthBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthBuf, uint32(len(bytesBuf)))
	if n, err := c.conn.Write(lengthBuf); err != nil {
		logrus.Error("Failed to send msg length; err: %v", err)
		return n, err
	}
	return c.conn.Write(bytesBuf)
}

func (c *ReplyChannel) GetSourceAddress() string {
	return c.localAddress
}

func (c *ReplyChannel) GetDestAddress() string {
	return c.destAddress
}

func NewTransport(port int, seedHost string, nodeId string) *Transport {
	var hostAddress string
	address, err := net.InterfaceAddrs()
	if err != nil {
		logrus.Fatal("Fail to get interfaces")
	}

	for _, a := range address {
		if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				hostAddress = ipNet.IP.String() + ":"
				hostAddress += strconv.Itoa(port)
				break
			}
		}
	}

	var seedHosts []string
	if len(seedHost) > 0 {
		seedHosts = strings.Split(seedHost, ",")
	}

	return &Transport{
		LocalAddress:    hostAddress,
		LocalNodeId:     nodeId,
		SeedHosts:       seedHosts,
		RequestHandlers: make(map[string]transport.RequestHandler),
	}
}

func (t *Transport) Register(action string, handler transport.RequestHandler) {
	t.RequestHandlers[action] = handler
}

func (t *Transport) Start(port int) {
	listen := "0.0.0.0:" + strconv.Itoa(port)
	go func() {
		l, err := net.Listen("tcp", listen)
		if err != nil {
			logrus.Fatalf("Fail to bind address to %s; err: %v", err)
		}
		logrus.Infof("Success of listening on %s", listen)
		defer l.Close()

		for {
			conn, err := l.Accept()
			if err != nil {
				logrus.Errorf("Fail to accept; err: %v", err)
				continue
			}
			go func(conn net.Conn) {
				for {
					lengthBuf := make([]byte, 4)
					if _, err := conn.Read(lengthBuf); err != nil {
						if io.EOF == err {
							logrus.Warnf("Connection is closed from client; %v", conn.RemoteAddr().String())
							return
						}
						logrus.Errorf("Fail to receive data; err: %v", err)
						return
					}
					msgLength := binary.LittleEndian.Uint32(lengthBuf)
					recvBuf := make([]byte, int(msgLength))
					n, err := io.ReadFull(conn, recvBuf)
					if err != nil {
						logrus.Errorf("Fail to get response; err: %v", err)
						return
					}
					if 0 < n {
						// Receive request data
						recvData := dataFormatFromBytes(recvBuf)
						action := recvData.Action
						data := recvData.Content

						if strings.Contains(action, "_FAIL") {
							logrus.Fatalln("Error: %s", string(data))
						} else {
							t.RequestHandlers[action](&ReplyChannel{
								requestId:    recvData.Id,
								conn:         conn,
								localAddress: t.LocalAddress,
								destAddress:  recvData.Source,
							}, data)
						}
					}
				}
			}(conn)
		}
	}()
}

func (t *Transport) OpenConnection(address string, callback func(conn transport.Connection)) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		logrus.Errorf("Failed to connect to %s : %v", address, err)
		callback(&Connection{
			err: "Failed to connect to " + address,
		})
		return
	}

	logrus.Info("Success on connecting ", address)

	c := &Connection{
		conn:             conn,
		localAddress:     t.LocalAddress,
		destAddress:      address,
		responseHandlers: map[uint64]func(byte []byte){},
	}
	callback(c)
}

func (t *Transport) GetLocalAddress() string {
	return t.LocalAddress
}

func (t *Transport) GetSeedHosts() []string {
	return t.SeedHosts
}

func (t *Transport) GetNodeId() string {
	return t.LocalNodeId
}

func (t *Transport) GetHandler(action string) transport.RequestHandler {
	return t.RequestHandlers[action]
}

// Data Format
type DataFormat struct {
	Id      uint64
	Source  string
	Dest    string
	Action  string
	Content []byte
}

func (d *DataFormat) toBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(d); err != nil {
		logrus.Fatalln(err)
	}
	return buffer.Bytes()
}

func dataFormatFromBytes(b []byte) *DataFormat {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data DataFormat
	if err := decoder.Decode(&data); err != nil {
		logrus.Fatalln(err)
	}
	return &data
}
