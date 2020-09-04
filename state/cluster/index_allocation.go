package cluster

import (
	"fmt"
	"github.com/actumn/searchgoose/state"
	"math"
)

func weight(node state.RoutingNode, index string) float64 {
	theta0 := 0.55
	theta1 := 0.45

	weightShard := float64(node.NumShards())             // - avgShardsPerNode
	weightIndex := float64(node.NumShardsOfIndex(index)) // - avgShardsPerNodeOfIndex
	return theta0*weightShard + theta1*weightIndex
}

type AllocationService struct {
}

func NewAllocationService() *AllocationService {
	return &AllocationService{}
}

// shard 마다 node 별 weight 계산, 앎맞는 data node id 를 할당.
func (s *AllocationService) reroute(clusterState state.ClusterState) state.ClusterState {
	routingNodes := state.NewRoutingNodes(clusterState)

	//nodes := len(clusterState.Nodes.DataNodes)
	//avgShardsPerNode := float64(0) / float64(nodes)
	for _, shard := range routingNodes.UnassignedShards {
		// decide to allocate unassigned
		var minNode *state.RoutingNode = nil

		minWeight := math.MaxFloat64
		for _, node := range routingNodes.NodesToShards {
			indexName := shard.ShardId.Index.Name
			//avgShardsPerNodeOfIndex := float64(clusterState.Metadata.Indices[indexName].RoutingNumShards / nodes)
			currentWeight := weight(*node, indexName)

			if currentWeight > minWeight {
				continue
			} else {
				minNode = node
				minWeight = currentWeight
			}
		}

		if minNode != nil {
			shard.CurrentNodeId = minNode.NodeId
			minNode.Add(*shard)
		}
	}
	fmt.Println()

	// generate routing table based on routing nodes
	newRoutingTable := state.RoutingTable{
		IndicesRouting: map[string]state.IndexRoutingTable{},
	}
	for _, node := range routingNodes.NodesToShards {
		for _, shard := range node.Shards {
			indexName := shard.ShardId.Index.Name
			if _, ok := newRoutingTable.IndicesRouting[indexName]; !ok {
				newRoutingTable.IndicesRouting[indexName] = state.IndexRoutingTable{
					Index:  shard.ShardId.Index,
					Shards: map[int]state.IndexShardRoutingTable{},
				}
			}

			newRoutingTable.IndicesRouting[indexName].Shards[shard.ShardId.ShardId] = state.IndexShardRoutingTable{
				ShardId: shard.ShardId,
				Primary: shard,
			}
		}
	}

	return state.ClusterState{
		Version:      clusterState.Version,
		StateUUID:    clusterState.StateUUID,
		Name:         clusterState.Name,
		Nodes:        clusterState.Nodes,
		Metadata:     clusterState.Metadata,
		RoutingTable: newRoutingTable,
	}
}
