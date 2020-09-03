package tcp

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"
	"net"
	"time"
)

type Transport struct {
	LocalAddress   string
	LocalNodeId    string
	SeedHosts      map[string]string    // SeedHostAddress -> NodeId
	ConnectedNodes map[string]*net.Conn // nodeId -> outbound connection (outbound nodes에서 관리 )
}

func NewTransport(address string, nodeId string, seedHosts []string) *Transport {
	hostMap := map[string]string{}
	for _, host := range seedHosts {
		hostMap[host] = ""
	}
	return &Transport{
		LocalAddress:   address,
		LocalNodeId:    nodeId,
		SeedHosts:      hostMap,
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
					// TODO :: 만약 target node로부터 ACK가 먼저 들어온다면?
					data := DataFormatFromBytes(recvBuf[:n])

					// log.Printf("Recived ACK from %s", string(data))
					// TODO :: 여기서 connection은 inbound handling만. (다른 노드로부터 데이터 수신시 이 connection을 사용, server role connection)
					if data.Action == HANDSHAKE_REQ {
						// REQ를 받았는데, seed_hosts 들과의 outbound connection이 아직 안 맺어져 있을 수 있다
						// 그렇다면, 내가 REQ를 받았다는 걸 저장하고 (SeedHost에 nodeId를 저장),
						// OpenConnection에서 seedhost에 저장되어 있으면(이전에 REQ 받았다는 뜻이므로) ACK 보내고, 아니면 REQ 보낸다
						destAddr := data.Source
						destNodeId := data.Content
						t.SeedHosts[destAddr] = string(destNodeId)
						log.Printf("Receive REQ from %s\n", destNodeId)

						if t.ConnectedNodes[destAddr] != nil {
							log.Printf("Already connected with %s\n", destNodeId)
							handShakeData := DataFormat{
								Source:  t.LocalAddress,
								Action:  HANDSHAKE_ACK,
								Content: []byte(t.LocalNodeId),
							}
							t.RequestHandShake(*t.ConnectedNodes[destAddr], handShakeData.ToBytes())
						}
					} else if data.Action == HANDSHAKE_ACK {
						// ACK를 받는다는 소리는, 내가 다른 노드에 REQ를 보냈다는 뜻이다
						// REQ를 보냈다는 소리는 나한테는 지금 seed_hosts 들과의 outbound connection이 맺어져 있다는 뜻

						addr := data.Source
						conn := t.ConnectedNodes[addr]
						delete(t.ConnectedNodes, addr)
						nodeId := string(data.Content)
						t.ConnectedNodes[nodeId] = conn

						log.Printf("Finished Handshake with %s\n", nodeId)

					} else if data.Action == HANDSHAKE_FAIL {

					}

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
	// TODO :: 여기서 만들어지는 connection으로 outbound handling. (다른 노드로 데이터 전송시 이 connection을 사용, client role connection)

	handShakeData := DataFormat{
		Source:  t.LocalAddress,
		Action:  HANDSHAKE_REQ,
		Content: []byte(t.LocalNodeId),
	}

	t.ConnectedNodes[destAddress] = &conn
	whereSend := destAddress

	if t.SeedHosts[destAddress] != "" {
		// 내가 다른 노드로부터 REQ를 받았다는 소리에요
		// 그럼 나는 outbound conection 맺고 ACK 만 쏴주면 되겠다.
		nodeId := t.SeedHosts[destAddress]
		whereSend = nodeId
		log.Printf("Already receive REQ from %s\n", nodeId)

		t.ConnectedNodes[nodeId] = &conn
		handShakeData.Action = HANDSHAKE_ACK
	}

	msg := "ACK"
	if handShakeData.Action == 0 {
		msg = "REQ"
	}

	log.Printf("Send %s to %s\n", msg, whereSend)
	if msg == "ACK" {
		log.Printf("Finished Handshake with %s\n", whereSend)
	}

	// TODO :: ACK를 받을 때 까지 계속 보내? 아니지 1분 정도는 connection 시도를 하자
	go t.RequestHandShake(conn, handShakeData.ToBytes())
	time.Sleep(time.Duration(10) * time.Second)

}

func (t *Transport) RequestHandShake(conn net.Conn, message []byte) {
	// TODO :: handshake패킷임을 알리는 비트플래그 또는 field 필요.
	for {
		//conn.Write([]byte("ping"))
		conn.Write(message)
		time.Sleep(time.Duration(3) * time.Second)
	}
}

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
