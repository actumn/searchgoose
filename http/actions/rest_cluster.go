package actions

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/sirupsen/logrus"
	"strconv"
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

	indicesInfo := map[string]interface{}{
		".kibana_task_manager_2": map[string]interface{}{ // TODO :: remove it
			"state": "open",
		},
	}
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

type RestClusterStats struct {
	clusterService   *cluster.Service
	transportService *transport.Service
}

func NewRestClusterStats(clusterService *cluster.Service, transportService *transport.Service, indicesService *indices.Service) *RestClusterStats {
	transportService.RegisterRequestHandler(ClusterStatsAction, func(channel transport.ReplyChannel, req []byte) {

	})

	return &RestClusterStats{
		clusterService:   clusterService,
		transportService: transportService,
	}
}

func (h *RestClusterStats) Handle(r *RestRequest, reply ResponseListener) {
	// TODO :: resolve nodes map, indices map from cluster state and broadcasting
	clusterState := h.clusterService.State()
	nodes := clusterState.Nodes

	indicesMap := map[string]interface{}{
		"count":  1,
		"shards": map[string]interface{}{},
		"docs": map[string]interface{}{
			"count":   1,
			"deleted": 0,
		},
		"mappings": map[string]interface{}{
			"field_types": []interface{}{},
		},
	}
	nodesMap := map[string]interface{}{
		"count": map[string]interface{}{
			"total":  1,
			"master": 1,
			"data":   1,
		},
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
			"indices":      indicesMap,
			"nodes":        nodesMap,
		},
	})
}
