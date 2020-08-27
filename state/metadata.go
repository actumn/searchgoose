package state

import (
	"bytes"
	"encoding/gob"
	"log"
)

// ClusterState
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

var (
	EmptyMetadata = Metadata{}
)

// Metadata
type IndexMetadataState int

const (
	OPEN IndexMetadataState = iota
	CLOSE
)

type Metadata struct {
	ClusterUUID string
	Version     int64
	//Coordination CoordinationMetadata
	Indices map[string]IndexMetadata
	//Templates    map[string]IndexTemplateMetadata
}

type Index struct {
	Name string
	Uuid string
}

type IndexMetadata struct {
	Index              Index
	RoutingNumShards   int
	RoutingNumReplicas int
	Version            int64
	State              IndexMetadataState
	Mapping            map[string]MappingMetadata
	//Settings Settings
}

/*
	type: "_doc"
	source: "properties: {}"
*/
type MappingMetadata struct {
	Type   string
	Source []byte
}

//type IndexTemplateMetadata struct {
//}
//
//type CoordinationMetadata struct {
//}
