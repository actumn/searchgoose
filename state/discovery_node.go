package state

import (
	"bytes"
	"encoding/gob"
	"github.com/sirupsen/logrus"
)

type Node struct {
	Name string
	Id   string
	//ephemeralId string
	//HostName    string
	HostAddress string
	//Address     Address
	//Attributes map[string]string
	//version Version
	//roles map[DiscoveryNodeRole]struct{}
}

func CreateLocalNode(id string, address string, name string) *Node {
	if name == "" {
		name = id
	}
	return &Node{
		Name:        name,
		Id:          id,
		HostAddress: address,
	}
}

func (n *Node) ToBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(n); err != nil {
		logrus.Fatalln(err)
	}
	return buffer.Bytes()
}

func NodeFromBytes(b []byte) *Node {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var node Node
	if err := decoder.Decode(&node); err != nil {
		logrus.Fatalln(err)
	}
	return &node
}

func isMasterNode() bool {
	return true
}

func isDataNode() bool {
	return true
}

type Nodes struct {
	Nodes        map[string]Node
	DataNodes    map[string]Node
	MasterNodes  map[string]Node
	MasterNodeId string
	LocalNodeId  string
}

func (n *Nodes) MasterNode() Node {
	return n.Nodes[n.MasterNodeId]
}
