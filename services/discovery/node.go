package discovery

type Node struct {
	Name        string
	Id          string
	ephemeralId string
	HostName    string
	HostAddress string
	//Address TransportAddress
	Attributess map[string]string
	//version Version
	//roles map[DiscoveryNodeRole]struct{}
}

func CreateLocal(name string, id string) {

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
