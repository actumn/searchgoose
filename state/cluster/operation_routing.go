package cluster

import (
	"github.com/actumn/searchgoose/common"
	"github.com/actumn/searchgoose/state"
)

func indexShards(clusterState state.ClusterState, index string, id string) state.IndexShardRoutingTable {
	indexMetadata := clusterState.Metadata.Indices[index]
	shardId := common.MurMur3Hash(id) % indexMetadata.RoutingNumShards
	return clusterState.RoutingTable.IndicesRouting[index].Shards[shardId]
}

func getShards() {

}
