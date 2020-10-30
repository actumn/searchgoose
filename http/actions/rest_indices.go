package actions

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
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

		channel.SendMessage(IndicesStatsAction, indicesStatsRes.toBytes())
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

		for _, shardStats := range response.ShardStats {
			if indexStats, existing := indicesStats[shardStats.ShardRouting.ShardId.Index.Name]; existing {
				indicesStats[shardStats.ShardRouting.ShardId.Index.Name] = index.Stats{
					Name:       shardStats.ShardRouting.ShardId.Index.Name,
					Uuid:       shardStats.ShardRouting.ShardId.Index.Uuid,
					ShardStats: append(indexStats.ShardStats, shardStats),
				}
			} else {
				indicesStats[shardStats.ShardRouting.ShardId.Index.Name] = index.Stats{
					Name:       shardStats.ShardRouting.ShardId.Index.Name,
					Uuid:       shardStats.ShardRouting.ShardId.Index.Uuid,
					ShardStats: []index.ShardStats{shardStats},
				}
			}
		}
	}

	indicesMap := map[string]interface{}{}
	allDocsCount := uint64(0)
	allDocsDeleted := uint64(0)
	allSizesInBytes := uint64(0)
	for indexName, indexStats := range indicesStats {
		docsCount := uint64(0)
		docsDeleted := uint64(0)
		sizesInBytes := uint64(0)
		for _, shardStats := range indexStats.ShardStats {
			docsCount += shardStats.NumDocs
			docsDeleted += shardStats.UserData["deletes"].(uint64)
			sizesInBytes += shardStats.UserData["num_bytes_used_disk"].(uint64)
		}
		allDocsCount += docsCount
		allDocsDeleted += docsDeleted
		allSizesInBytes += sizesInBytes

		indicesMap[indexName] = map[string]interface{}{
			"uuid": indexStats.Uuid,
			"primaries": map[string]interface{}{
				"docs": map[string]interface{}{
					"count":   docsCount,
					"deleted": docsDeleted,
				},
				"store": map[string]interface{}{
					"size_in_bytes": sizesInBytes,
				},
				"indexing": map[string]interface{}{
					"index_total":           docsCount,
					"index_time_in_millis":  -1,
					"index_current":         0,
					"index_failed":          0,
					"delete_total":          docsDeleted,
					"delete_time_in_millis": -1,
					"delete_current":        0,
				},
			},
			"total": map[string]interface{}{
				"docs": map[string]interface{}{
					"count":   docsCount,
					"deleted": docsDeleted,
				},
				"store": map[string]interface{}{
					"size_in_bytes": sizesInBytes,
				},
				"indexing": map[string]interface{}{
					"index_total":           docsCount,
					"index_time_in_millis":  -1,
					"index_current":         0,
					"index_failed":          0,
					"delete_total":          docsDeleted,
					"delete_time_in_millis": -1,
					"delete_current":        0,
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
			"_all": map[string]interface{}{
				"primaries": map[string]interface{}{
					"docs": map[string]interface{}{
						"count":   allDocsCount,
						"deleted": allDocsDeleted,
					},
					"store": map[string]interface{}{
						"size_in_bytes": allSizesInBytes,
					},
					"indexing": map[string]interface{}{
						"index_total":           allDocsCount,
						"index_time_in_millis":  -1,
						"index_current":         0,
						"index_failed":          0,
						"delete_total":          allDocsDeleted,
						"delete_time_in_millis": -1,
						"delete_current":        0,
					},
				},
				"total": map[string]interface{}{
					"docs": map[string]interface{}{
						"count":   allDocsCount,
						"deleted": allDocsDeleted,
					},
					"store": map[string]interface{}{
						"size_in_bytes": allSizesInBytes,
					},
					"indexing": map[string]interface{}{
						"index_total":           allDocsCount,
						"index_time_in_millis":  -1,
						"index_current":         0,
						"index_failed":          0,
						"delete_total":          allDocsDeleted,
						"delete_time_in_millis": -1,
						"delete_current":        0,
					},
				},
			},
			"indices": indicesMap,
		},
	})
}

type RestGetMappings struct {
	clusterService              *cluster.Service
	indexNameExpressionResolver *indices.NameExpressionResolver
}

func NewRestGetMappings(clusterService *cluster.Service, indexNameExpressionResolver *indices.NameExpressionResolver) *RestGetMappings {
	return &RestGetMappings{
		clusterService:              clusterService,
		indexNameExpressionResolver: indexNameExpressionResolver,
	}
}

func (h *RestGetMappings) Handle(r *RestRequest, reply ResponseListener) {
	indexExpression := r.PathParams["index"]
	if indexExpression == "" {
		indexExpression = "*"
	}

	clusterState := h.clusterService.State()
	concreteIndices := h.indexNameExpressionResolver.ConcreteIndexNames(*clusterState, indexExpression)

	indicesInfo := map[string]interface{}{}
	for _, indexName := range concreteIndices {
		indexMetadata := clusterState.Metadata.Indices[indexName]

		var mappings map[string]interface{}
		if err := json.Unmarshal(indexMetadata.Mapping["_doc"].Source, &mappings); err != nil {
			logrus.Fatal(err)
		}
		indicesInfo[indexName] = map[string]interface{}{
			"mappings": mappings,
		}
	}
	reply(RestResponse{
		StatusCode: 200,
		Body:       indicesInfo,
	})
}
