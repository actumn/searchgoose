package services

type ClusterState struct {
	Version   int64
	stateUUID string
	Name      string
	Metadata  Metadata
	Blocks    Blocks
	Nodes     *Nodes
}

type Blocks struct {
}

type CoordinationState struct {
	LocalNode      *Node
	PersistedState PersistedState
}
