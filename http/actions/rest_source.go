package actions

import (
	"errors"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/sirupsen/logrus"
)

type RestGetSource struct {
	clusterService              *cluster.Service
	indicesService              *indices.Service
	indexNameExpressionResolver *indices.NameExpressionResolver
	transportService            *transport.Service
}

func NewRestGetSource(clusterService *cluster.Service, indicesService *indices.Service, indexNameExpressionResolver *indices.NameExpressionResolver, transportService *transport.Service) *RestGetSource {
	return &RestGetSource{
		clusterService:              clusterService,
		indicesService:              indicesService,
		indexNameExpressionResolver: indexNameExpressionResolver,
		transportService:            transportService,
	}
}

func (h *RestGetSource) Handle(r *RestRequest, reply ResponseListener) {
	indexExpression := r.PathParams["index"]
	documentId := r.PathParams["id"]

	clusterState := h.clusterService.State()
	indexName := h.indexNameExpressionResolver.ConcreteSingleIndex(*clusterState, indexExpression).Name
	if indexName == "" {
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

	shardRouting := cluster.GetShards(*clusterState, indexName, documentId).Primary
	getRequest := getRequest{
		Index:   indexName,
		Id:      documentId,
		ShardId: shardRouting.ShardId,
	}
	h.transportService.SendRequest(*clusterState.Nodes.Nodes[shardRouting.CurrentNodeId], GetAction, getRequest.toBytes(), func(response []byte) {
		res := getResponseFromBytes(response)
		if res.Err != "" {
			logrus.Warn(errors.New(res.Err))
			reply(RestResponse{
				StatusCode: 400,
				Body: map[string]interface{}{
					"_index": indexName,
					"_type":  "_doc",
					"_id":    documentId,
					"found":  res.Err,
				},
			})
		} else {
			reply(RestResponse{
				StatusCode: 200,
				Body:       res.Fields,
			})
		}
	})
}
