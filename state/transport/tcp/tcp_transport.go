package tcp

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"
	"net"
)

type Transport struct {
	LocalAddress   string
	LocalNodeId    string
	SeedHosts      []string
	ConnectedNodes map[string]*net.Conn // nodeId -> outbound connection (outbound nodes에서 관리 )
}

func NewTransport(address string, nodeId string, seedHosts []string) *Transport {
	return &Transport{
		LocalAddress:   address,
		LocalNodeId:    nodeId,
		SeedHosts:      seedHosts,
		ConnectedNodes: map[string]*net.Conn{},
	}
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
					// Receive REQ
					data := DataFormatFromBytes(recvBuf[:n])
					log.Printf("Receive REQ from %s\n", data.Source)
					handShakeData := DataFormat{
						Source:  t.LocalAddress,
						Action:  HANDSHAKE_ACK,
						Content: []byte(t.LocalNodeId),
					}
					message := handShakeData.ToBytes()
					// Send ACK
					log.Printf("Send ACK to %s\n", data.Source)
					conn.Write(message)
				}
			}(conn)
		}
	}()
}

func (t *Transport) OpenConnection(destAddress string, c chan *net.Conn) {
	conn, err := net.Dial("tcp", destAddress)
	if err != nil {
		log.Fatalf("Failed to connect to %s : %v", destAddress, err)
	}
	log.Printf("Success on connecting %s\n", destAddress)

	handShakeData := DataFormat{
		Source:  t.LocalAddress,
		Action:  HANDSHAKE_REQ,
		Content: []byte(t.LocalNodeId),
	}
	message := handShakeData.ToBytes()

	// TODO :: ACK를 받을 때 까지 계속 보내? 아니지 1분 정도는 connection 시도를 하자

	// Send REQ
	log.Printf("Send REQ to %s\n", destAddress)
	conn.Write(message)

	// Wait ACK
	recvBuf := make([]byte, 4096)
	n, err := conn.Read(recvBuf)
	if err != nil {
		log.Fatalf("Fail to get ACK from %s; err: %v", destAddress, err)
		return
	}

	data := DataFormatFromBytes(recvBuf[:n])
	nodeId := string(data.Content)
	log.Printf("Receive ACK from %s\n", string(data.Content))

	t.ConnectedNodes[nodeId] = &conn

	log.Printf("Finished handshaking with %s\n", nodeId)
}

// TODO :: Action type string으로 바꾸기
type Action int

const (
	HANDSHAKE_REQ Action = iota
	HANDSHAKE_ACK
	HANDSHAKE_FAIL
)

type DataFormat struct {
	Source string
	//Dest   string
	Action  Action
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
