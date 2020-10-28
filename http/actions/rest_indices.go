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

type indicesStatsRequest struct {
	NodeId string
	Shards []state.ShardRouting
}

func (r *indicesStatsRequest) toBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func indicesStatsRequestFromBytes(b []byte) *indicesStatsRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var req indicesStatsRequest
	if err := decoder.Decode(&req); err != nil {
		logrus.Fatal(err)
	}
	return &req
}

type indicesStatsResponse struct {
	TotalShards int
	ShardStats  []index.ShardStats
}

func (r *indicesStatsResponse) toBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func indicesStatsResponseFromBytes(b []byte) *indicesStatsResponse {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var req indicesStatsResponse
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
		indicesStatsReq := indicesStatsRequestFromBytes(req)
		var shardStats []index.ShardStats

		for _, shardRouting := range indicesStatsReq.Shards {
			indexService := indicesService.Indices[shardRouting.ShardId.Index.Uuid]
			indexShard, _ := indexService.Shard(shardRouting.ShardId.ShardId)
			shardStats = append(shardStats, indexShard.Stats())
		}

		indicesStatsRes := indicesStatsResponse{
			TotalShards: len(indicesStatsReq.Shards),
			ShardStats:  shardStats,
		}

		channel.SendMessage("", indicesStatsRes.toBytes())
	})

	return &RestIndicesStatsAction{
		clusterService:              clusterService,
		indicesService:              indicesService,
		indexNameExpressionResolver: indexNameExpressionResolver,
		transportService:            transportService,
	}
}

func (h *RestIndicesStatsAction) Handle(r *RestRequest, reply ResponseListener) {
	indexExpression := r.PathParams["index"]
	if indexExpression == "" {
		indexExpression = "*"
	}
	clusterState := h.clusterService.State()

	concreteIndices := h.indexNameExpressionResolver.ConcreteIndexNames(*clusterState, indexExpression)
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

	totalShards := 0
	indicesStats := map[string]index.Stats{}
	for _, response := range responses {
		totalShards += response.TotalShards

		for _, shardStat := range response.ShardStats {
			if indexStats, existing := indicesStats[shardStat.ShardRouting.ShardId.Index.Name]; existing {
				indicesStats[shardStat.ShardRouting.ShardId.Index.Name] = index.Stats{
					Name:       shardStat.ShardRouting.ShardId.Index.Name,
					Uuid:       shardStat.ShardRouting.ShardId.Index.Uuid,
					ShardStats: append(indexStats.ShardStats, shardStat),
				}
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
