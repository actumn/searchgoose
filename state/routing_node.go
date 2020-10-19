package state

type RoutingNodes struct {
	NodesToShards    map[string]*RoutingNode
	UnassignedShards []*ShardRouting
}

func NewRoutingNodes(clusterState ClusterState) *RoutingNodes {
	nodesToShards := map[string]*RoutingNode{}
	var unassignedShards []*ShardRouting

	for id, node := range clusterState.Nodes.DataNodes {
		nodesToShards[id] = &RoutingNode{
			NodeId:        id,
			node:          &node,
			Shards:        map[ShardId]ShardRouting{},
			ShardsByIndex: map[string]map[ShardRouting]struct{}{},
		}
	}

	for _, indexRoutingTable := range clusterState.RoutingTable.IndicesRouting {
		for _, shardRoutingTable := range indexRoutingTable.Shards {
			primary := shardRoutingTable.Primary
			if primary.CurrentNodeId != "" {
				node := nodesToShards[primary.CurrentNodeId]
				node.Add(primary)
			} else {
				unassignedShards = append(unassignedShards, &primary)
			}
		}
	}

	return &RoutingNodes{
		NodesToShards:    nodesToShards,
		UnassignedShards: unassignedShards,
	}
}

type RoutingNode struct {
	NodeId        string
	node          *Node
	Shards        map[ShardId]ShardRouting
	ShardsByIndex map[string]map[ShardRouting]struct{}
}

func (n *RoutingNode) Add(shard ShardRouting) {
	n.Shards[shard.ShardId] = shard
	if _, ok := n.ShardsByIndex[shard.ShardId.Index.Name]; !ok {
		n.ShardsByIndex[shard.ShardId.Index.Name] = map[ShardRouting]struct{}{}
	}
	n.ShardsByIndex[shard.ShardId.Index.Name][shard] = struct{}{}
}

func (n *RoutingNode) NumShards() int {
	return len(n.Shards)
}
func (n *RoutingNode) NumShardsOfIndex(index string) int {
	return len(n.ShardsByIndex[index])
}
