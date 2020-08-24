package state

import (
	"bytes"
	"encoding/gob"
	"log"
)

type ClusterService interface {
	State() *ClusterState
	SubmitStateUpdateTask(task ClusterStateUpdateTask)
}

type ClusterStateUpdateTask func(s ClusterState) ClusterState

type ClusterChangedEvent struct {
	State     ClusterState
	PrevState ClusterState
}

type ClusterState struct {
	Version   int64
	StateUUID string
	Name      string
	Nodes     *Nodes
	Metadata  Metadata
	//Blocks    ClusterBlocks
	//RoutingTable RoutingTable
}

func (c *ClusterState) ToBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(c); err != nil {
		log.Fatalln(err)
	}
	return buffer.Bytes()
}
func ClusterStateFromBytes(b []byte, localNode *Node) *ClusterState {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var state ClusterState
	if err := decoder.Decode(&state); err != nil {
		log.Fatalln(err)
	}

	state.Nodes.LocalNodeId = localNode.Id
	state.Nodes.Nodes[localNode.Id] = localNode
	return &state
}

type ClusterBlocks struct {
}

type CoordinationState struct {
	LocalNode      *Node
	PersistedState PersistedState
}
