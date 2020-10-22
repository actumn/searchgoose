package actions

import (
	"encoding/json"
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type RestGetIndex struct {
	clusterService              *cluster.Service
	indexNameExpressionResolver *indices.NameExpressionResolver
}

func NewRestGetIndex(clusterService *cluster.Service, indexNameExpressionResolver *indices.NameExpressionResolver) *RestGetIndex {
	return &RestGetIndex{
		clusterService:              clusterService,
		indexNameExpressionResolver: indexNameExpressionResolver,
	}
}

func (h *RestGetIndex) Handle(r *RestRequest, reply ResponseListener) {
	// TODO:: forward to master if local node is data node
	//indicesExpressions := strings.Split(, ",")
	indexExpression := r.PathParams["index"]
	if indexExpression == "_all" {
		indexExpression = "*"
	}
	clusterState := h.clusterService.State()
	indexNames := h.indexNameExpressionResolver.ConcreteIndexNames(*clusterState, indexExpression)
	if !strings.Contains(indexExpression, "*") && len(indexNames) == 0 {
		reply(RestResponse{
			StatusCode: 404,
			Body: map[string]interface{}{
				"error": map[string]interface{}{
					"root_cause": []map[string]interface{}{
						{
							"type":          "index_not_found_exception",
							"reason":        "no such index [" + indexExpression + "]",
							"resource.type": "index_or_alias",
							"resource.id":   ".kibana",
							"index_uuid":    "_na_",
							"index":         ".kibana",
						},
					},
					"type":          "index_not_found_exception",
					"reason":        "no such index [" + indexExpression + "]",
					"resource.type": "index_or_alias",
					"resource.id":   indexExpression,
					"index_uuid":    "_na_",
					"index":         indexExpression,
				},
				"status": 404,
			},
		})
		return
	}

	response := map[string]interface{}{}
	for _, indexName := range indexNames {
		index := clusterState.Metadata.Indices[indexName]
		var mappings map[string]interface{}
		if err := json.Unmarshal(index.Mapping["_doc"].Source, &mappings); err != nil {
			logrus.Fatal(err)
		}

		aliases := map[string]interface{}{}
		for _, alias := range index.Aliases {
			aliases[alias.Alias] = map[string]interface{}{}
		}
		response[indexName] = map[string]interface{}{
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
		Body:       response,
	})
}

type RestPutIndex struct {
	createIndexService *cluster.MetadataCreateIndexService
}

func NewRestPutIndex(clusterIndexService *cluster.MetadataCreateIndexService) *RestPutIndex {
	return &RestPutIndex{
		createIndexService: clusterIndexService,
	}
}

func (h *RestPutIndex) Handle(r *RestRequest, reply ResponseListener) {
	index := r.PathParams["index"]

	var body map[string]interface{}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		reply(RestResponse{
			StatusCode: 400,
			Body: map[string]interface{}{
				"err": err,
			},
		})
		return
	}

	mapping, err := json.Marshal(body["mappings"])
	if err != nil {
		reply(RestResponse{
			StatusCode: 400,
			Body: map[string]interface{}{
				"err": err,
			},
		})
		return
	}

	req := cluster.CreateIndexClusterStateUpdateRequest{
		Index:    index,
		Mappings: mapping,
		Settings: body["settings"].(map[string]interface{}),
	}
	h.createIndexService.CreateIndex(req)

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"acknowledged":        true,
			"shards_acknowledged": true,
			"index":               index,
		},
	})
}

type RestDeleteIndex struct {
	clusterService              *cluster.Service
	indexNameExpressionResolver *indices.NameExpressionResolver
	deleteIndexService          *cluster.MetadataDeleteIndexService
}

func NewRestDeleteIndex(clusterService *cluster.Service, indexNameExpressionResolver *indices.NameExpressionResolver, deleteIndexService *cluster.MetadataDeleteIndexService) *RestDeleteIndex {
	return &RestDeleteIndex{
		clusterService:              clusterService,
		indexNameExpressionResolver: indexNameExpressionResolver,
		deleteIndexService:          deleteIndexService,
	}
}

func (h *RestDeleteIndex) Handle(r *RestRequest, reply ResponseListener) {
	indexExpression := r.PathParams["index"]
	clusterState := h.clusterService.State()
	indexNames := h.indexNameExpressionResolver.ConcreteIndexNames(*clusterState, indexExpression)

	req := cluster.DeleteIndexClusterStateUpdateRequest{
		Indices: []state.Index{},
	}
	for _, indexName := range indexNames {
		req.Indices = append(req.Indices, clusterState.Metadata.Indices[indexName].Index)
	}

	h.deleteIndexService.DeleteIndex(req)

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"acknowledged": true,
		},
	})
}

type RestHeadIndex struct {
	clusterService              *cluster.Service
	indexNameExpressionResolver *indices.NameExpressionResolver
}

func NewRestHeadIndex(clusterService *cluster.Service, indexNameExpressionResolver *indices.NameExpressionResolver) *RestHeadIndex {
	return &RestHeadIndex{
		clusterService:              clusterService,
		indexNameExpressionResolver: indexNameExpressionResolver,
	}
}

func (h *RestHeadIndex) Handle(r *RestRequest, reply ResponseListener) {
	// TODO:: forward to master if local node is data node
	//indicesExpressions := strings.Split(, ",")
	indexExpression := r.PathParams["index"]
	clusterState := h.clusterService.State()
	indexNames := h.indexNameExpressionResolver.ConcreteIndexNames(*clusterState, indexExpression)
	if !strings.Contains(indexExpression, "*") && len(indexNames) == 0 {
		reply(RestResponse{
			StatusCode: 404,
			Body:       "",
		})
		return
	}

	reply(RestResponse{
		StatusCode: 200,
		Body:       "",
	})
}
