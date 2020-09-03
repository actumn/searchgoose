package tcp

import (
	"io"
	"log"
	"net"
	"time"
)

type Transport struct {
	//localConnection *net.Conn
	LocalNodeId    string
	ConnectedNodes map[string]*net.Conn // nodeId -> Connection
}

func NewTransport(nodeId string) *Transport {
	return &Transport{
		LocalNodeId:    nodeId,
		ConnectedNodes: map[string]*net.Conn{},
	}
}

func (t *Transport) Start(srcAddress string) {
	go func() {
		l, err := net.Listen("tcp", ":"+srcAddress)
		// l, err := net.Listen("tcp", ":8180")
		if err != nil {
			log.Fatalf("Fail to bind address to %s; err: %v", srcAddress, err)
		}
		log.Printf("Success of listening on %s\n", srcAddress)
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
					data := recvBuf[:n]
					log.Printf("Recived ACK from %s", string(data))
					// TODO :: 여기서 connection은 inbound handling만. (다른 노드로부터 데이터 수신시 이 connection을 사용, server role connection)

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
	// 서로 nodeId 교환 필요.
	// remoteNodeId := string(data)
	// t.ConnectedNodes[remoteNodeId] = &conn

	connectedNodes := make(chan string)
	// TODO :: ACK를 받을 때 까지 계속 보내? 아니지 1분 정도는 connection 시도를 하자
	go t.HandShake(conn, connectedNodes)
	time.Sleep(time.Duration(10) * time.Second)
	//c <- &conn

}

func (t *Transport) HandShake(conn net.Conn, connectedNodes chan<- string) {
	// TODO :: handshake패킷임을 알리는 비트플래그 또는 field 필요.
	for {
		//conn.Write([]byte("ping"))
		conn.Write([]byte(t.LocalNodeId))
		time.Sleep(time.Duration(3) * time.Second)
	}
}

type OpenConnectionRequest struct {
	NodeId string
}

type OpenConnectionResponse struct {
	NodeId string
}
