package discovery

import "github.com/actumn/searchgoose/services/transport"

type Node struct {
	Name        string
	Id          string
	ephemeralId string
	HostName    string
	HostAddress string
	Address     transport.Address
	Attributess map[string]string
	//version Version
	//roles map[DiscoveryNodeRole]struct{}
}

func CreateLocal(id string) *Node {
	return &Node{
		Name: "testName",
		Id:   id,
	}
}

func isMasterNode() bool {
	return true
}

func isDataNode() bool {
	return true
}

type Nodes struct {
	Nodes        map[string]*Node
	DataNodes    map[string]*Node
	MasterNodes  map[string]*Node
	MasterNodeId string
	LocalNodeId  string
}
