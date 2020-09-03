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
				// TODO :: 여기서 어떻게 main에 있는 ConnectedNodes에 연결시키지?
				// remoteNodeId := string(data)
				// t.ConnectedNodes[remoteNodeId] = &conn

			}
		}(conn)
	}
}

func (t *Transport) Send(destAddress string, message []byte) {
	conn, err := net.Dial("tcp", destAddress)
	if err != nil {
		log.Fatalf("Failed to connect to %s : %v", destAddress, err)
	}

	for {
		//conn.Write([]byte("ping"))
		conn.Write(message)
		time.Sleep(time.Duration(3) * time.Second)
	}
}

func (t *Transport) OpenConnection(destAddress string, c chan *net.Conn) {
	conn, err := net.Dial("tcp", destAddress)
	if err != nil {
		log.Fatalf("Failed to connect to %s : %v", destAddress, err)
	}
	log.Printf("Success on connecting %s\n", destAddress)
	connectedNodes := make(chan string)
	// TODO :: ACK를 받을 때 까지 계속 보내? 아니지 1분 정도는 connection 시도를 하자
	go t.HandShake(conn, connectedNodes)
	time.Sleep(time.Duration(10) * time.Second)
	//c <- &conn

}

func (t *Transport) HandShake(conn net.Conn, connectedNodes chan<- string) {
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
