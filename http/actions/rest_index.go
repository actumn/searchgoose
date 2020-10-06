package actions

import (
	"encoding/json"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"log"
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
	clusterState := h.clusterService.State()
	indexNames := h.indexNameExpressionResolver.ConcreteIndexNames(*clusterState, indexExpression)
	if !strings.Contains(indexExpression, "*") && len(indexNames) == 0 {
		reply(RestResponse{
			StatusCode: 200,
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
			log.Fatalln(err)
		}

		response[indexName] = map[string]interface{}{
			"aliases":  map[string]interface{}{},
			"mappings": mappings,
			"settings": map[string]interface{}{
				"index": map[string]interface{}{
					"creation_date":      "1597382566866",
					"number_of_shards":   strconv.Itoa(index.RoutingNumShards),
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
	CreateIndexService *cluster.MetadataCreateIndexService
}

func NewRestPutIndex(clusterIndexService *cluster.MetadataCreateIndexService) *RestPutIndex {
	return &RestPutIndex{
		CreateIndexService: clusterIndexService,
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

	h.CreateIndexService.CreateIndex(struct {
		Index    string
		Mappings []byte
	}{Index: index, Mappings: mapping})

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
}

func NewRestDeleteIndex() *RestDeleteIndex {
	return &RestDeleteIndex{}
}

func (h *RestDeleteIndex) Handle(r *RestRequest, reply ResponseListener) {
	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"acknowledged": true,
		},
	})
}

type RestHeadIndex struct{}

func (h *RestHeadIndex) Handle(r *RestRequest, reply ResponseListener) {
	reply(RestResponse{
		StatusCode: 200,
	})
}
