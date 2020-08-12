package tcp

import (
	"io"
	"log"
	"net"
	"time"
)

type Transport struct {
}

func (t *Transport) Start() {
	l, err := net.Listen("tcp", ":8180")
	if err != nil {
		log.Fatalf("Fail to bind address to 8180; err: %v", err)
	}
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
				data := recvBuf[:n]
				log.Println(string(data))
			}
		}(conn)
	}
}

func (t *Transport) Send() {
	conn, err := net.Dial("tcp", "lcoalhost:8180")
	if err != nil {
		log.Fatalf("Failed to connect to server", err)
	}

	for {
		conn.Write([]byte("ping"))
		time.Sleep(time.Duration(3) * time.Second)
	}
}
