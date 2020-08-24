package state

type Node struct {
	Name        string
	Id          string
	ephemeralId string
	HostName    string
	HostAddress string
	//Address     Address
	Attributes map[string]string
	//version Version
	//roles map[DiscoveryNodeRole]struct{}
}

func CreateLocalNode(id string) *Node {
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

func (n *Nodes) MasterNode() *Node {
	return n.Nodes[n.MasterNodeId]
}

type RoutingTable struct {
}
