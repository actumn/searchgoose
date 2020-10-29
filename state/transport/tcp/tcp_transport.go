package tcp

import (
	"bytes"
	"encoding/gob"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"strconv"
	"strings"
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
	err          string
}

func (c *Connection) SendRequest(action string, content []byte, callback func(byte []byte)) {

	request := DataFormat{
		Source:  c.GetSourceAddress(),
		Dest:    c.GetDestAddress(),
		Action:  action,
		Content: content,
	}
	logrus.Infof("Send %s to %s\n", request.Action, request.Dest)

	_, err := c.conn.Write(request.ToBytes())
	if err != nil {
		logrus.Infof("Fail to send request; err:%v\n", err)
	}
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, c.conn)
		if err != nil {
			// logrus.Fatalf("Fail to get response from %s; err: %v", address, err)
			logrus.Warnf("Fail to get response; err: %v", err)
			return
		}
		response := DataFormatFromBytes(buf.Bytes())
		logrus.Infof("Receive %s from %s\n", response.Action, response.Source)
		if strings.Contains(response.Action, "_FAIL") {
			logrus.Warnf("%s", string(response.Content))
		} else {
			callback(response.Content)
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
				logrus.Infof("Fail to accept; err: %v", err)
				continue
			}
			go func(conn net.Conn) {
				for {
					var buf bytes.Buffer
					n, err := io.Copy(&buf, conn)
					if err != nil {
						if io.EOF == err {
							logrus.Infof("Connection is closed from client; %v", conn.RemoteAddr().String())
							return
						}
						logrus.Infof("Fail to receive data; err: %v", err)
						return
					}
					if 0 < n {
						// Receive request data
						recvData := DataFormatFromBytes(buf.Bytes())
						action := recvData.Action
						data := recvData.Content

						if strings.Contains(action, "_FAIL") {
							logrus.Fatalln("Error: %s", string(data))
						} else {
							t.RequestHandlers[action](&ReplyChannel{
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
		logrus.Warnf("Failed to connect to %s : %v", address, err)
		callback(&Connection{
			err: "Failed to connect to " + address,
		})
		return
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

func (t *Transport) GetNodeId() string {
	return t.LocalNodeId
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
		logrus.Fatalln(err)
	}
	return buffer.Bytes()
}

func DataFormatFromBytes(b []byte) *DataFormat {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data DataFormat
	if err := decoder.Decode(&data); err != nil {
		logrus.Fatalln(err)
	}
	return &data
}
