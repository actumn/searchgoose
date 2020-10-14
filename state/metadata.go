package state

import (
	"bytes"
	"encoding/gob"
	"github.com/actumn/searchgoose/common"
	"github.com/sirupsen/logrus"
)

// ClusterState
type ClusterState struct {
	Version   int64
	StateUUID string
	Name      string
	Nodes     *Nodes
	Metadata  Metadata
	//Blocks    ClusterBlocks
	RoutingTable RoutingTable
}

func (c *ClusterState) ToBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(c); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}
func ClusterStateFromBytes(b []byte, localNode *Node) *ClusterState {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var state ClusterState
	if err := decoder.Decode(&state); err != nil {
		logrus.Fatal(err)
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

	Term int64 // 원래는 PersistedState에서 관리, 임시로 CoordinationState 여기서 관리하도록

	JoinVotes    *VoteCollection
	PublishVotes *VoteCollection
	ElectionWon  bool
}

func (c *CoordinationState) getCurrentTerm() int64 {
	return c.Term
}

func (c *CoordinationState) IsElectionQuorum(nodes []string) bool {
	return c.JoinVotes.IsQuorum(nodes)
}

// VoteCollection
type VoteCollection struct {
	nodes map[string]*Node
	joins map[Join]bool //Set data structure
}

func NewVoteCollection() *VoteCollection {
	return &VoteCollection{
		nodes: make(map[string]*Node),
		joins: make(map[Join]bool),
	}
}

func (v *VoteCollection) IsQuorum(nodes []string) bool {
	votes := make([]string, 0, len(v.nodes))

	for key := range v.nodes {
		votes = append(votes, key)
	}

	intersection := common.GetIntersection(nodes, votes)
	logrus.Info(77, "configured", len(nodes), "votes", len(votes), "intersection", len(intersection))

	return len(intersection)*2 > len(nodes)
}

func (v *VoteCollection) AddVote(sourceNode *Node) bool {
	// master node 인지 체크하기
	for k, _ := range v.nodes {
		if k == sourceNode.Id {
			v.nodes[sourceNode.Id] = sourceNode
			return false
		}
	}
	v.nodes[sourceNode.Id] = sourceNode
	return true
}

func (v *VoteCollection) AddJoinVote(join Join) bool {
	added := v.AddVote(&(join.SourceNode))
	if added {
		// Set data structure
		v.joins[join] = true
	}
	return added
}

type Join struct {
	SourceNode Node
	TargetNode Node
	Term       int64
}

func NewJoin(sourceNode Node, targetNode Node, term int64) *Join {
	return &Join{
		SourceNode: sourceNode,
		TargetNode: targetNode,
		Term:       term,
	}
}

var (
	EmptyMetadata = Metadata{
		Indices: map[string]IndexMetadata{},
	}
)

// Metadata
type IndexMetadataState int

const (
	OPEN IndexMetadataState = iota
	CLOSE
)

type Metadata struct {
	//ClusterUUID string
	//Version     int64
	//Coordination CoordinationMetadata
	Indices map[string]IndexMetadata
	//Templates    map[string]IndexTemplateMetadata
}

type Index struct {
	Name string
	Uuid string
}

type IndexMetadata struct {
	Index            Index
	RoutingNumShards int
	//RoutingNumReplicas int
	//Version            int64
	//State              IndexMetadataState
	Aliases map[string]AliasMetadata
	Mapping map[string]MappingMetadata
	//Settings Settings
}

type AliasMetadata struct {
	Alias string
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

type VotingConfiguration struct {
	// nodeIds  map[string]bool
	NodeIds []string
}

func NewVotingConfiguration() *VotingConfiguration {
	return &VotingConfiguration{
		NodeIds: []string{},
	}
}

type RoutingTable struct {
	IndicesRouting map[string]IndexRoutingTable
}

type IndexRoutingTable struct {
	Index  Index
	Shards map[int]IndexShardRoutingTable
}

type IndexShardRoutingTable struct {
	ShardId ShardId
	Primary ShardRouting
	//Replicas []ShardRouting
}

type ShardId struct {
	Index   Index
	ShardId int
}

type ShardRouting struct {
	ShardId       ShardId
	CurrentNodeId string
	//RelocatingNodeId string
	Primary bool
}
