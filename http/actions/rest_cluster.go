package actions

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/actumn/searchgoose/index"
	"github.com/actumn/searchgoose/monitor"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/sirupsen/logrus"
	"strconv"
	"sync"
)

const (
	ClusterStatsAction = "cluster:monitor/stats"
)

type RestClusterHealth struct {
	clusterService              *cluster.Service
	indexNameExpressionResolver *indices.NameExpressionResolver
}

func NewRestClusterHealth(clusterService *cluster.Service, indexNameExpressionResolver *indices.NameExpressionResolver) *RestClusterHealth {
	return &RestClusterHealth{
		clusterService:              clusterService,
		indexNameExpressionResolver: indexNameExpressionResolver,
	}
}

func (h *RestClusterHealth) Handle(r *RestRequest, reply ResponseListener) {
	clusterState := h.clusterService.State()
	indicesNames := h.indexNameExpressionResolver.ConcreteIndexNames(*clusterState, "*")
	activePrimaryShards, activeShards := 0, 0
	for _, indexName := range indicesNames {
		//indexMetadata := clusterState.Metadata.Indices[indexName]
		indexRoutingTable := clusterState.RoutingTable.IndicesRouting[indexName]

		activePrimaryShards += len(indexRoutingTable.Shards)
		activeShards += len(indexRoutingTable.Shards)
	}

	nodes := clusterState.Nodes
	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"cluster_name":                     clusterState.Name,
			"status":                           "green",
			"timed_out":                        false,
			"number_of_nodes":                  len(nodes.Nodes),
			"number_of_data_nodes":             len(nodes.DataNodes),
			"active_primary_shards":            activePrimaryShards,
			"active_shards":                    activeShards,
			"relocating_shards":                0,
			"initializing_shards":              0,
			"unassigned_shards":                0,
			"delayed_unassigned_shards":        0,
			"number_of_pending_tasks":          0,
			"number_of_in_flight_fetch":        0,
			"task_max_waiting_in_queue_millis": 0,
			"active_shards_percent_as_number":  100,
		},
	})
}

type RestClusterState struct {
	clusterService *cluster.Service
}

func NewRestClusterStateMetadata(clusterService *cluster.Service) *RestClusterState {
	return &RestClusterState{
		clusterService: clusterService,
	}
}

func (h *RestClusterState) Handle(r *RestRequest, reply ResponseListener) {
	clusterState := h.clusterService.State()
	metadata := clusterState.Metadata

	indicesInfo := map[string]interface{}{}
	for _, index := range metadata.Indices {
		var mappings map[string]interface{}
		if err := json.Unmarshal(index.Mapping["_doc"].Source, &mappings); err != nil {
			logrus.Fatal(err)
		}

		aliases := map[string]interface{}{}
		for _, alias := range index.Aliases {
			aliases[alias.Alias] = map[string]interface{}{}
		}
		indicesInfo[index.Index.Name] = map[string]interface{}{
			"state":    "open",
			"aliases":  aliases,
			"mappings": mappings,
			"settings": map[string]interface{}{
				"index": map[string]interface{}{
					"creation_date":      "1597382566866",
					"number_of_shards":   strconv.Itoa(index.NumberOfShards),
					"number_of_replicas": "0",
					"uuid":               index.Index.Uuid,
					"version": map[string]interface{}{
						"created": "7080299",
					},
					"provided_name": index.Index.Name,
				},
			},
		}
	}

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"cluster_name": clusterState.Name,
			"cluster_uuid": clusterState.StateUUID,
			"metadata": map[string]interface{}{
				"cluster_uuid":           clusterState.StateUUID,
				"cluster_uuid_committed": true,
				"templates":              map[string]interface{}{},
				"indices":                indicesInfo,
				"index_lifecycle":        map[string]interface{}{},
				"index-graveyard":        map[string]interface{}{},
				"ingest":                 map[string]interface{}{},
			},
		},
	})
}

type clusterStatsNodeRequest struct {
}

func (r *clusterStatsNodeRequest) toBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}
func clusterStatsNodeRequestFromBytes(b []byte) *clusterStatsNodeRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var req clusterStatsNodeRequest
	if err := decoder.Decode(&req); err != nil {
		logrus.Fatal(err)
	}
	return &req
}

type clusterStatsNodeResponse struct {
	NodeStats  monitor.Stats
	ShardStats []index.ShardStats
}

func (r *clusterStatsNodeResponse) toBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}
func clusterStatsNodeResponseFromBytes(b []byte) *clusterStatsNodeResponse {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var req clusterStatsNodeResponse
	if err := decoder.Decode(&req); err != nil {
		logrus.Fatal(err)
	}
	return &req
}

type RestClusterStats struct {
	clusterService   *cluster.Service
	transportService *transport.Service
}

func NewRestClusterStats(clusterService *cluster.Service, transportService *transport.Service, indicesService *indices.Service) *RestClusterStats {
	monitorService := monitor.NewService()

	transportService.RegisterRequestHandler(ClusterStatsAction, func(channel transport.ReplyChannel, req []byte) {
		nodeStats := monitorService.Stats()
		var shardStats []index.ShardStats
		for _, indexService := range indicesService.Indices {
			for _, shard := range indexService.Shards {
				shardStats = append(shardStats, shard.Stats())
			}
		}

		res := clusterStatsNodeResponse{
			NodeStats:  nodeStats,
			ShardStats: shardStats,
		}
		channel.SendMessage("", res.toBytes())
	})

	return &RestClusterStats{
		clusterService:   clusterService,
		transportService: transportService,
	}
}

func (h *RestClusterStats) Handle(r *RestRequest, reply ResponseListener) {
	clusterState := h.clusterService.State()
	nodes := clusterState.Nodes

	responses := make([]clusterStatsNodeResponse, len(nodes.Nodes))
	wg := sync.WaitGroup{}
	wg.Add(len(nodes.Nodes))
	idx := -1
	for _, node := range nodes.Nodes {
		idx += 1
		currIdx := idx
		h.transportService.SendRequest(node, ClusterStatsAction, []byte(""), func(response []byte) {
			res := clusterStatsNodeResponseFromBytes(response)
			responses[currIdx] = *res
			wg.Done()
		})
	}
	wg.Wait()

	memTotal := uint64(0)
	memFree := uint64(0)

	fsTotal := uint64(0)
	fsFree := uint64(0)
	fsAvailable := uint64(0)

	indicesCount := map[string]struct{}{}
	shards := 0
	primaries := 0
	docs := uint64(0)
	numBytesUsedDisk := uint64(0)
	for _, response := range responses {
		memTotal += response.NodeStats.Os.Mem.Total
		memFree += response.NodeStats.Os.Mem.Free

		fsTotal += response.NodeStats.Fs.Total
		fsFree += response.NodeStats.Fs.Free
		fsAvailable += response.NodeStats.Fs.Available

		for _, shardStat := range response.ShardStats {
			docs += shardStat.NumDocs

			indicesCount[shardStat.ShardRouting.ShardId.Index.Name] = struct{}{}
			shards += 1
			if shardStat.ShardRouting.Primary {
				primaries += 1
			}
			numBytesUsedDisk += shardStat.UserData["num_bytes_used_disk"].(uint64)
		}
	}

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"_nodes": map[string]interface{}{
				"total":      len(nodes.Nodes),
				"successful": len(nodes.Nodes),
				"failed":     0,
			},
			"cluster_name": clusterState.Name,
			"cluster_uuid": clusterState.StateUUID,
			"status":       "green",
			"indices": map[string]interface{}{
				"count": len(indicesCount),
				"shards": map[string]interface{}{
					"total":     shards,
					"primaries": primaries,
				},
				"docs": map[string]interface{}{
					"count":   docs,
					"deleted": -1,
				},
				"store": map[string]interface{}{
					"size_in_bytes": numBytesUsedDisk,
				},
			},
			"nodes": map[string]interface{}{
				"count": map[string]interface{}{
					"total":  len(nodes.Nodes),
					"master": 1,
					"data":   len(nodes.DataNodes),
				},
				"os": map[string]interface{}{
					"mem": map[string]interface{}{
						"total_in_bytes": memTotal,
						"free_in_bytes":  memFree,
						"used_in_bytes":  memTotal - memFree,
						"free_percent":   memFree * 100 / memTotal,
						"used_percent":   100 - memFree*100/memTotal,
					},
				},
				"fs": map[string]interface{}{
					"total_in_bytes":     fsTotal,
					"free_in_bytes":      fsFree,
					"available_in_bytes": fsAvailable,
				},
			},
		},
	})
}
