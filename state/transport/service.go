package transport

import (
	"bytes"
	"encoding/gob"
	"github.com/actumn/searchgoose/state"
	"github.com/sirupsen/logrus"
	"sync"
)

const (
	HANDSHAKE_REQ  = "HANDSHAKE_REQ"
	HANDSHAKE_ACK  = "HANDSHAKE_ACK"
	HANDSHAKE_FAIL = "HANDSHAKE_FAIL"
	PEERFIND_REQ   = "PEERFIND_REQ"
	PEERFIND_ACK   = "PEERFIND_ACK"
	PEERFIND_FAIL  = "PEERFIND_FAIL"
	PREVOTE_REQ    = "PREVOTE_REQ"
	PREVOTE_RES    = "PREVOTE_RES"
	START_JOIN     = "START_JOIN"
	START_JOIN_ACK = "START_JOIN_ACK"
	JOIN_REQ       = "JOIN_REQ"
)

// Interfaces
type Connection interface {
	SendRequest(action string, req []byte, callback func(byte []byte))
	GetSourceAddress() string
	GetDestAddress() string
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

func NewService(id string, transport Transport) *Service {
	address := transport.GetLocalAddress()
	service := &Service{
		LocalNode:         state.CreateLocalNode(id, address),
		Transport:         transport,
		ConnectionManager: make(map[string]ConnectionEntry),
		ConnectionLock:    sync.RWMutex{},
	}
	service.RegisterRequestHandler(HANDSHAKE_REQ, service.handleHandshake)
	service.RegisterRequestHandler(PEERFIND_REQ, service.handlePeersRequest)

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
		logrus.Printf("ConnectToRemoteNode(%s) not connecting local node ", address)
		return
	}

	for _, value := range s.ConnectionManager {
		if value.node.HostAddress == address {
			logrus.Printf("Connection is already established; %s", address)
			return
		}
	}

	// TODO :: goroutine 으로 빼면 좋을 것 같다
	//var mutex = &sync.Mutex{}
	s.Transport.OpenConnection(address, func(conn Connection) {
		handshakeData := HandshakeRequest{
			RemoteAddress: address,
		}
		content := handshakeData.ToBytes()

		s.SendRequestConn(conn, HANDSHAKE_REQ, content, func(res []byte) {
			data := HandshakeResponseFromBytes(res)
			logrus.Info("Success on handshaking with %v\n", data.Node)

			connectedNode := data.Node

			s.ConnectionManager[connectedNode.Id] = ConnectionEntry{
				conn: conn,
				node: connectedNode,
			}

			/*
				s.SetConnection(connectedNode.Id, ConnectionEntry{
					conn: conn,
					node: connectedNode,
				})

			*/

			callback(&connectedNode)
		})
	})
}

func (s *Service) RequestPeers(node state.Node, knownPeers []state.Node, callback func()) {
	nowNode := *(s.LocalNode)
	peerFindData := PeersRequest{
		SourceNode: nowNode,
		KnownPeers: knownPeers,
	}

	logrus.Printf("Peer=%v requesting peers %s\n", nowNode, knownPeers)

	// TODO :: 나중에 request handler interface로 뽑아내기
	request := peerFindData.ToBytes()

	s.SendRequest(node, PEERFIND_REQ, request, func(res []byte) {
		data := PeersResponseFromBytes(res)
		peers := data.KnownPeers

		logrus.Printf("Peer=%v received PeersResponse=%v\n", nowNode, data)

		for _, peer := range peers {
			go s.ConnectToRemoteNode(peer.HostAddress, func(remoteNode *state.Node) {
				_, connectedNodes := s.GetConnectedPeers()
				s.RequestPeers(*remoteNode, connectedNodes, func() {})
			})
		}
		callback()
	})
}

func (s *Service) IsConnected(address string) bool {
	if _, ok := s.ConnectionManager[address]; ok {
		return true
	}
	return false
}

func (s *Service) GetLocalNode() state.Node {
	return *(s.LocalNode)
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

func (s *Service) handlePeersRequest(channel ReplyChannel, req []byte) {

	peerReqData := PeersRequestFromBytes(req)
	logrus.Printf("Receive Peer Finding REQ from %s; %s\n", channel.GetDestAddress(), peerReqData.KnownPeers)

	peers := peerReqData.KnownPeers
	for _, peer := range peers {
		go s.ConnectToRemoteNode(peer.HostAddress, func(remoteNode *state.Node) {
			_, connectedNodes := s.GetConnectedPeers()
			s.RequestPeers(*remoteNode, connectedNodes, func() {})
		})
	}

	knownPeers := make([]state.Node, 0, len(s.ConnectionManager))
	for _, peer := range s.ConnectionManager {
		knownPeers = append(knownPeers, peer.node)
	}

	logrus.Info("Send Peer Finding RES to %s; %s\n", channel.GetDestAddress(), knownPeers)

	peerResData := PeersResponse{
		KnownPeers: knownPeers,
	}

	response := peerResData.ToBytes()
	channel.SendMessage(PEERFIND_ACK, response)
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

// Peer-find
type PeersRequest struct {
	SourceNode state.Node
	KnownPeers []state.Node
}

func (r *PeersRequest) ToBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func PeersRequestFromBytes(b []byte) *PeersRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data PeersRequest
	if err := decoder.Decode(&data); err != nil {
		logrus.Fatal(err)
	}
	return &data
}

type PeersResponse struct {
	MasterNode state.Node
	KnownPeers []state.Node
}

func (r *PeersResponse) ToBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func PeersResponseFromBytes(b []byte) *PeersResponse {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data PeersResponse
	if err := decoder.Decode(&data); err != nil {
		logrus.Fatal(err)
	}
	return &data
}

type LocalConnection struct {
	service *Service
}

func (c *LocalConnection) SendRequest(action string, req []byte, callback func(byte []byte)) {
	handler := c.service.Transport.GetHandler(action)
	replyChannel := &DirectReplyChannel{
		callback: callback,
	}

	handler(replyChannel, req)
}

func (c *LocalConnection) GetDestAddress() string {
	return c.service.LocalNode.HostAddress
}

func (c *LocalConnection) GetSourceAddress() string {
	return c.service.LocalNode.HostAddress
}

type DirectReplyChannel struct {
	// address string
	callback func(byte []byte)
}

func (c *DirectReplyChannel) SendMessage(action string, b []byte) (int, error) {
	c.callback(b)
	return 0, nil
}

func (c *DirectReplyChannel) GetDestAddress() string {
	return c.GetDestAddress()
}

func (c *DirectReplyChannel) GetSourceAddress() string {
	return c.GetSourceAddress()
}
