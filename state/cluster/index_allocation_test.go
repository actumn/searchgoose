package cluster

import (
	"fmt"
	"github.com/actumn/searchgoose/state"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAllocationService_reroute(t *testing.T) {
	// Arrange
	allocationService := NewAllocationService()
	clusterState := state.ClusterState{
		Version:   0,
		StateUUID: "testUUID",
		Name:      "test",
		Nodes: &state.Nodes{
			DataNodes: map[string]*state.Node{
				"testNodeId1": {
					Name:        "node1",
					Id:          "testNodeId1",
					HostAddress: "",
				},
				"testNodeId2": {
					Name:        "node2",
					Id:          "testNodeId2",
					HostAddress: "",
				},

				"testNodeId3": {
					Name:        "node3",
					Id:          "testNodeId3",
					HostAddress: "",
				},
			},
		},
		Metadata: state.Metadata{
			Indices: map[string]state.IndexMetadata{
				"test": {
					Index: state.Index{
						Name: "test",
						Uuid: "testUuid",
					},
					RoutingNumShards: 3,
				},
			},
		},
		RoutingTable: state.RoutingTable{
			IndicesRouting: map[string]state.IndexRoutingTable{
				"test": {
					Index: state.Index{
						Name: "test",
						Uuid: "testUuid",
					},
					Shards: map[int]state.IndexShardRoutingTable{
						0: {
							ShardId: state.ShardId{
								Index: state.Index{
									Name: "test",
									Uuid: "testUuid",
								},
								ShardId: 0,
							},
							Primary: state.ShardRouting{
								ShardId: state.ShardId{
									Index: state.Index{
										Name: "test",
										Uuid: "testUuid",
									},
									ShardId: 0,
								},
								CurrentNodeId: "",
								Primary:       true,
							},
						},
						1: {
							ShardId: state.ShardId{
								Index: state.Index{
									Name: "test",
									Uuid: "testUuid",
								},
								ShardId: 1,
							},
							Primary: state.ShardRouting{
								ShardId: state.ShardId{
									Index: state.Index{
										Name: "test",
										Uuid: "testUuid",
									},
									ShardId: 1,
								},
								CurrentNodeId: "",
								Primary:       true,
							},
						},
						2: {
							ShardId: state.ShardId{
								Index: state.Index{
									Name: "test",
									Uuid: "testUuid",
								},
								ShardId: 2,
							},
							Primary: state.ShardRouting{
								ShardId: state.ShardId{
									Index: state.Index{
										Name: "test",
										Uuid: "testUuid",
									},
									ShardId: 2,
								},
								CurrentNodeId: "",
								Primary:       true,
							},
						},
					},
				},
			},
		},
	}

	// Action
	result := allocationService.reroute(clusterState)

	// Assert
	fmt.Println(result.RoutingTable.IndicesRouting)
	assert.NotEqual(t, "", result.RoutingTable.IndicesRouting["test"].Shards[0].Primary.CurrentNodeId)
}
