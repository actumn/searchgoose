package actions

import (
	"bytes"
	"encoding/gob"
	"github.com/actumn/searchgoose/index"
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/sirupsen/logrus"
	"sync"
)

const (
	IndicesStatsAction = "indices:monitor/stats"
)

type nodeRequest struct {
	NodeId string
	Shards []state.ShardRouting
}

func (r *nodeRequest) toBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func nodeRequestFromBytes(b []byte) *nodeRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var req nodeRequest
	if err := decoder.Decode(&req); err != nil {
		logrus.Fatal(err)
	}
	return &req
}

type nodeResponse struct {
	TotalShards int
	ShardStats  []index.ShardStats
}

func (r *nodeResponse) toBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func nodeResponseFromBytes(b []byte) *nodeResponse {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var req nodeResponse
	if err := decoder.Decode(&req); err != nil {
		logrus.Fatal(err)
	}
	return &req
}

type RestIndicesStatsAction struct {
	clusterService              *cluster.Service
	indicesService              *indices.Service
	indexNameExpressionResolver *indices.NameExpressionResolver
	transportService            *transport.Service
}

func NewRestIndicesStatsAction(clusterService *cluster.Service, indicesService *indices.Service, indexNameExpressionResolver *indices.NameExpressionResolver, transportService *transport.Service) *RestIndicesStatsAction {
	transportService.RegisterRequestHandler(IndicesStatsAction, func(channel transport.ReplyChannel, req []byte) {
		nodeReq := nodeRequestFromBytes(req)
		var shardStats []index.ShardStats

		for _, shardRouting := range nodeReq.Shards {
			indexService := indicesService.Indices[shardRouting.ShardId.Index.Uuid]
			indexShard, _ := indexService.Shard(shardRouting.ShardId.ShardId)
			shardStats = append(shardStats, indexShard.Stats())
		}

		nodeRes := nodeResponse{
			TotalShards: len(nodeReq.Shards),
			ShardStats:  shardStats,
		}

		channel.SendMessage("", nodeRes.toBytes())
	})

	return &RestIndicesStatsAction{
		clusterService:              clusterService,
		indicesService:              indicesService,
		indexNameExpressionResolver: indexNameExpressionResolver,
		transportService:            transportService,
	}
}

func (h *RestIndicesStatsAction) Handle(r *RestRequest, reply ResponseListener) {
	// TODO:: reply based on indices service and all cluster shards information
	clusterState := h.clusterService.State()

	concreteIndices := h.indexNameExpressionResolver.ConcreteIndexNames(*clusterState, "*")
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

	if len(nodeIds) == 0 {
		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"_shards": map[string]interface{}{
					"total":      0,
					"successful": 0,
					"failed":     0,
				},
				"_all": map[string]interface{}{
					"primaries": map[string]interface{}{},
					"total":     map[string]interface{}{},
				},
				"indices": map[string]interface{}{},
			},
		})
		return
	}

	responses := make([]nodeResponse, len(nodeIds))
	wg := sync.WaitGroup{}
	wg.Add(len(nodeIds))
	idx := -1
	for nodeId, shards := range nodeIds {
		idx += 1
		node := clusterState.Nodes.Nodes[nodeId]
		nodeReq := nodeRequest{
			NodeId: nodeId,
			Shards: shards,
		}
		func(idx int) {
			h.transportService.SendRequest(node, IndicesStatsAction, nodeReq.toBytes(), func(response []byte) {
				nodeRes := nodeResponseFromBytes(response)
				responses[idx] = *nodeRes
				wg.Done()
			})
		}(idx)
	}
	wg.Wait()

	totalShards := 0
	indicesStats := map[string]index.Stats{}
	for _, response := range responses {
		totalShards += response.TotalShards

		for _, shardStat := range response.ShardStats {
			if indexStats, existing := indicesStats[shardStat.ShardRouting.ShardId.Index.Name]; existing {
				indexStats.ShardStats = append(indexStats.ShardStats, shardStat)
			} else {
				indicesStats[shardStat.ShardRouting.ShardId.Index.Name] = index.Stats{
					Name:       shardStat.ShardRouting.ShardId.Index.Name,
					Uuid:       shardStat.ShardRouting.ShardId.Index.Uuid,
					ShardStats: []index.ShardStats{shardStat},
				}
			}
		}
	}

	indicesMap := map[string]interface{}{}
	for indexName, indexStats := range indicesStats {
		docsCount := uint64(0)
		sizesInBytes := uint64(0)
		for _, shardStat := range indexStats.ShardStats {
			docsCount += shardStat.NumDocs
			sizesInBytes += shardStat.UserData["num_bytes_used_disk"].(uint64)
		}

		indicesMap[indexName] = map[string]interface{}{
			"uuid": indexStats.Uuid,
			"primaries": map[string]interface{}{
				"docs": map[string]interface{}{
					"count":   docsCount,
					"deleted": -1,
				},
				"store": map[string]interface{}{
					"size_in_bytes": sizesInBytes,
				},
			},
		}
	}

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"_shards": map[string]interface{}{
				"total":      totalShards,
				"successful": totalShards,
				"failed":     0,
			},
			//"_all": map[string]interface{}{
			//	"primaries": map[string]interface{}{},
			//	"total":     map[string]interface{}{},
			//},
			"indices": indicesMap,
		},
	})
}
