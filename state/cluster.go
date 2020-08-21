package state

type ClusterService interface {
	State() *ClusterState
}

type ClusterState struct {
	Version   int64
	StateUUID string
	Name      string
	Nodes     *Nodes
	Metadata  Metadata
	//Blocks    Blocks
	//RoutingTable RoutingTable
}

type Blocks struct {
}

type CoordinationState struct {
	LocalNode      *Node
	PersistedState PersistedState
}
