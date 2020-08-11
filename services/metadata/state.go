package metadata

import "github.com/actumn/searchgoose/services/discovery"

type ClusterState struct {
	Version   int64
	stateUUID string
	Name      string
	Metadata  Metadata
	Blocks    Blocks
	Nodes     discovery.Nodes
}

type Blocks struct {
}

type CoordinationState struct {
}
