package actions

import (
	"github.com/actumn/searchgoose/common"
	"github.com/actumn/searchgoose/index"
	"github.com/actumn/searchgoose/monitor"
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/actumn/searchgoose/state/transport"
	"strconv"
	"sync"
)

type RestCatTemplates struct{}

func (h *RestCatTemplates) Handle(r *RestRequest, reply ResponseListener) {
	reply(RestResponse{
		StatusCode: 200,
		Body:       []interface{}{},
	})
}

type RestCatNodes struct {
	clusterService   *cluster.Service
	transportService *transport.Service
}

func NewRestCatNodes(clusterService *cluster.Service, transportService *transport.Service) *RestCatNodes {
	return &RestCatNodes{
		clusterService:   clusterService,
		transportService: transportService,
	}
}

func (h *RestCatNodes) Handle(r *RestRequest, reply ResponseListener) {
	clusterState := h.clusterService.State()

	nodes := clusterState.Nodes.Nodes
	responses := make([]nodeStatsResponse, len(nodes))
	wg := sync.WaitGroup{}
	wg.Add(len(nodes))
	idx := -1
	for _, node := range nodes {
		idx += 1
		currIdx := idx
		h.transportService.SendRequest(node, NodesStatsAction, []byte(""), func(response []byte) {
			nodeStatsRes := nodeStatsResponseFromBytes(response)
			responses[currIdx] = *nodeStatsRes
			wg.Done()
		})
	}
	wg.Wait()

	nodeStatsMap := map[string]monitor.Stats{}
	for _, response := range responses {
		nodeStatsMap[response.Node.Id] = response.NodeStats
	}

	var nodesList []map[string]interface{}
	for n, node := range clusterState.Nodes.Nodes {
		nodeStats := nodeStatsMap[node.Id]
		heapPer := nodeStats.Runtime.HeapAlloc * 100 / nodeStats.Runtime.HeapSys
		nodesList = append(nodesList, map[string]interface{}{
			"id":   node.Id,
			"m":    "*", // master
			"n":    n + node.Name,
			"u":    "44m",    // uptime
			"role": "dilmrt", // node role
			//"hc":         "156.8mb", // heap current
			//"hm":         "512mb",   // heap max
			"hp": strconv.FormatUint(heapPer, 10), // heap percent
			"ip": node.HostAddress,
			//"dt":         "468.4gb", // disk total
			//"du":         "267.4gb", // disk used
			"disk.avail": common.IBytes(nodeStats.Fs.Available), // disk available
			"l":          "-1",                                  //
		})
	}
	reply(RestResponse{
		StatusCode: 200,
		Body:       nodesList,
	})
}

type RestCatIndices struct {
	clusterService              *cluster.Service
	indexNameExpressionResolver *indices.NameExpressionResolver
	transportService            *transport.Service
}

func NewRestCatIndices(clusterService *cluster.Service, indexNameExpressionResolver *indices.NameExpressionResolver, transportService *transport.Service) *RestCatIndices {
	return &RestCatIndices{
		clusterService:              clusterService,
		indexNameExpressionResolver: indexNameExpressionResolver,
		transportService:            transportService,
	}
}

func (h *RestCatIndices) Handle(r *RestRequest, reply ResponseListener) {
	clusterState := h.clusterService.State()
	var indicesList []map[string]interface{}
	for _, indexMetadata := range clusterState.Metadata.Indices {
		// TODO :: resolve indices information from broadcasting
		indicesList = append(indicesList, map[string]interface{}{
			"health": "green",
			"status": "open",
			"index":  indexMetadata.Index.Name,
			"uuid":   indexMetadata.Index.Uuid,
			//"pri":            "1",
			//"rep":            "0",
			//"docs.count":     "100",
			//"docs.deleted":   "0",
			//"store.size":     "208b",
			//"pri.store.size": "208b",
		})
	}

	reply(RestResponse{
		StatusCode: 200,
		Body:       indicesList,
	})
}

type RestCatShards struct {
	clusterService              *cluster.Service
	indexNameExpressionResolver *indices.NameExpressionResolver
	transportService            *transport.Service
}

func NewRestCatShards(clusterService *cluster.Service, indexNameExpressionResolver *indices.NameExpressionResolver, transportService *transport.Service) *RestCatShards {
	return &RestCatShards{
		clusterService:              clusterService,
		indexNameExpressionResolver: indexNameExpressionResolver,
		transportService:            transportService,
	}
}

func (h *RestCatShards) Handle(r *RestRequest, reply ResponseListener) {
	indexExpression := r.PathParams["index"]
	if indexExpression == "" {
		indexExpression = "*"
	}

	clusterState := h.clusterService.State()
	concreteIndices := h.indexNameExpressionResolver.ConcreteIndexNames(*clusterState, indexExpression)
	if len(concreteIndices) == 0 {
		reply(RestResponse{
			StatusCode: 404,
			Body:       []interface{}{},
		})
		return
	}

	nodeIds := map[string][]state.ShardRouting{}
	for _, indexName := range concreteIndices {
		indexRoutingTable := clusterState.RoutingTable.IndicesRouting[indexName]
		for _, indexShardRoutingTable := range indexRoutingTable.Shards {
			shard := indexShardRoutingTable.Primary

			nodeId := shard.CurrentNodeId
			if shardList, existing := nodeIds[nodeId]; existing {
				nodeIds[nodeId] = append(shardList, shard)
			} else {
				nodeIds[nodeId] = []state.ShardRouting{shard}
			}
		}
	}

	responses := make([]indicesStatsResponse, len(nodeIds))
	wg := sync.WaitGroup{}
	wg.Add(len(nodeIds))
	idx := -1
	for nodeId, shards := range nodeIds {
		idx += 1
		node := clusterState.Nodes.Nodes[nodeId]
		indicesStatsReq := indicesStatsRequest{
			NodeId: nodeId,
			Shards: shards,
		}
		currIdx := idx
		h.transportService.SendRequest(node, IndicesStatsAction, indicesStatsReq.toBytes(), func(response []byte) {
			indicesStatsRes := indicesStatsResponseFromBytes(response)
			responses[currIdx] = *indicesStatsRes
			wg.Done()
		})
	}
	wg.Wait()

	shardsStats := map[state.ShardRouting]index.ShardStats{}
	for _, response := range responses {
		for _, shardStats := range response.ShardStats {
			shardsStats[shardStats.ShardRouting] = shardStats
		}
	}

	var shardsInfo []map[string]interface{}
	for _, indexRouting := range clusterState.RoutingTable.IndicesRouting {
		for _, shardRouting := range indexRouting.Shards {
			shardsInfo = append(shardsInfo, map[string]interface{}{
				"index":  shardRouting.ShardId.Index.Name,
				"shard":  shardRouting.ShardId.ShardId,
				"prirep": "p",
				"state":  "STARTED",
				"docs":   shardsStats[shardRouting.Primary].NumDocs,
				"store":  common.IBytes(shardsStats[shardRouting.Primary].UserData["num_bytes_used_disk"].(uint64)),
				"node":   shardRouting.Primary.CurrentNodeId,
			})
		}
	}

	reply(RestResponse{
		StatusCode: 200,
		Body:       shardsInfo,
	})
}
