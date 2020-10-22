package transport

import (
	"bytes"
	"encoding/gob"
	"github.com/actumn/searchgoose/state"
	"github.com/sirupsen/logrus"
	"sync"
)

const (
	HANDSHAKE_REQ   = "HANDSHAKE_REQ"
	HANDSHAKE_ACK   = "HANDSHAKE_ACK"
	HANDSHAKE_FAIL  = "HANDSHAKE_FAIL"
	PEERFIND_REQ    = "PEERFIND_REQ"
	PEERFIND_ACK    = "PEERFIND_ACK"
	PEERFIND_FAIL   = "PEERFIND_FAIL"
	PREVOTE_REQ     = "PREVOTE_REQ"
	PREVOTE_RES     = "PREVOTE_RES"
	PREVOTE_FAIL    = "PREVOTE_FAIL"
	START_JOIN_REQ  = "START_JOIN"
	START_JOIN_ACK  = "START_JOIN_ACK"
	START_JOIN_FAIL = "START_JOIN_FAIL"
	JOIN_REQ        = "JOIN_REQ"
	PUBLISH_REQ     = "PUBLISH_REQ"
	PUBLISH_ACK     = "PUBLISH_ACK"
)

// Interfaces
type Connection interface {
	SendRequest(action string, req []byte, callback func(byte []byte))
	GetSourceAddress() string
	GetDestAddress() string
	GetMessage() string
}

type ReplyChannel interface {
	SendMessage(action string, content []byte) (n int, err error)
	GetSourceAddress() string
	GetDestAddress() string
}

type Transport interface {
	OpenConnection(address string, callback func(conn Connection))
	Start(address string)
	Register(action string, handler RequestHandler)
	GetLocalAddress() string
	GetSeedHosts() []string
	GetNodeId() string
	GetHandler(action string) RequestHandler
}

// structures
type RequestHandler func(channel ReplyChannel, req []byte)

type ConnectionEntry struct {
	conn Connection
	node state.Node
}

type Service struct {
	LocalNode         *state.Node
	Transport         Transport
	ConnectionManager map[string]ConnectionEntry
	ConnectionLock    sync.RWMutex
}

func NewService(transport Transport) *Service {
	address := transport.GetLocalAddress()
	id := transport.GetNodeId()
	service := &Service{
		LocalNode:         state.CreateLocalNode(id, address),
		Transport:         transport,
		ConnectionManager: make(map[string]ConnectionEntry),
		ConnectionLock:    sync.RWMutex{},
	}
	service.RegisterRequestHandler(HANDSHAKE_REQ, service.handleHandshake)

	return service
}

func (s *Service) Start() {
	address := s.Transport.GetLocalAddress()
	s.Transport.Start(address)
}

func (s *Service) SendRequestConn(conn Connection, action string, req []byte, callback func(response []byte)) {
	conn.SendRequest(action, req, callback)
}

func (s *Service) SendRequest(node state.Node, action string, req []byte, callback func(response []byte)) {
	var conn Connection
	if node.Id == s.LocalNode.Id {
		conn = &LocalConnection{
			service: s,
		}
	} else {
		conn = s.ConnectionManager[node.Id].conn
	}

	s.SendRequestConn(conn, action, req, callback)
}

func (s *Service) RegisterRequestHandler(action string, handler RequestHandler) {
	s.Transport.Register(action, handler)
}

func (s *Service) GetConnection(id string) ConnectionEntry {
	s.ConnectionLock.RLock()
	defer s.ConnectionLock.Unlock()
	return s.ConnectionManager[id]
}

func (s *Service) SetConnection(key string, conn ConnectionEntry) {
	s.ConnectionLock.Lock()
	s.ConnectionManager[key] = conn
	s.ConnectionLock.Unlock()
}

func (s *Service) GetConnectedPeers() ([]string, []state.Node) {
	ids := make([]string, 0, len(s.ConnectionManager))
	values := make([]state.Node, 0, len(s.ConnectionManager))
	for _, v := range s.ConnectionManager {
		ids = append(ids, v.node.Id)
		values = append(values, v.node)
	}
	return ids, values
}

func (s *Service) ConnectToRemoteNode(address string, callback func(node *state.Node)) {

	curNode := s.LocalNode

	if curNode.HostAddress == address {
		logrus.Infof("ConnectToRemoteNode(%s) not connecting local node ", address)
		return
	}

	if node := s.GetNodeByAddress(address); node != nil {
		logrus.Infof("Connection is already established; %s", address)
		callback(node)
		return
	}

	// TODO :: goroutine 으로 빼면 좋을 것 같다
	s.Transport.OpenConnection(address, func(conn Connection) {
		if len(conn.GetMessage()) > 0 {
			callback(nil)
			return
		}

		handshakeData := HandshakeRequest{
			RemoteAddress: address,
		}
		content := handshakeData.ToBytes()

		s.SendRequestConn(conn, HANDSHAKE_REQ, content, func(res []byte) {
			data := HandshakeResponseFromBytes(res)
			logrus.Infof("Success on handshaking with %v\n", data.Node)

			connectedNode := data.Node

			s.SetConnection(connectedNode.Id, ConnectionEntry{
				conn: conn,
				node: connectedNode,
			})

			callback(&connectedNode)
		})
	})
}

func (s *Service) GetNodeByAddress(address string) *state.Node {
	for _, value := range s.ConnectionManager {
		if value.node.HostAddress == address {
			return &value.node
		}
	}
	return nil
}

func (s *Service) GetLocalNode() state.Node {
	return *(s.LocalNode)
}

func (s *Service) GetSeedHosts() []string {
	return s.Transport.GetSeedHosts()
}

// handlers

func (s *Service) handleHandshake(channel ReplyChannel, req []byte) {
	handShakeData := HandshakeResponse{
		Node: *s.LocalNode,
		// clusterName:
	}

	response := handShakeData.ToBytes()
	channel.SendMessage(HANDSHAKE_ACK, response)
}

// Templates

// Handshake
type HandshakeRequest struct {
	RemoteAddress string
}

func (h *HandshakeRequest) ToBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(h); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func HandshakeRequestFromBytes(b []byte) *HandshakeRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data HandshakeRequest
	if err := decoder.Decode(&data); err != nil {
		logrus.Fatal(err)
	}
	return &data
}

type HandshakeResponse struct {
	Node state.Node
	// ClusterName string
}

func (h *HandshakeResponse) ToBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(h); err != nil {
		logrus.Fatalln(err)
	}
	return buffer.Bytes()
}

func HandshakeResponseFromBytes(b []byte) *HandshakeResponse {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data HandshakeResponse
	if err := decoder.Decode(&data); err != nil {
		logrus.Fatalln(err)
	}
	return &data
}

type LocalConnection struct {
	service *Service
}

func (c *LocalConnection) SendRequest(action string, req []byte, callback func(response []byte)) {
	handler := c.service.Transport.GetHandler(action)
	replyChannel := &DirectReplyChannel{
		callback: callback,
		address:  c.service.LocalNode.HostAddress,
	}

	handler(replyChannel, req)
}

func (c *LocalConnection) GetDestAddress() string {
	return c.service.LocalNode.HostAddress
}

func (c *LocalConnection) GetSourceAddress() string {
	return c.service.LocalNode.HostAddress
}

func (c *LocalConnection) GetMessage() string {
	return ""
}

type DirectReplyChannel struct {
	address  string
	callback func(byte []byte)
}

func (c *DirectReplyChannel) SendMessage(action string, b []byte) (int, error) {
	c.callback(b)
	return 0, nil
}

func (c *DirectReplyChannel) GetDestAddress() string {
	return c.address
}

func (c *DirectReplyChannel) GetSourceAddress() string {
	return c.address
}
