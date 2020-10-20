package actions

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"github.com/actumn/searchgoose/common"
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/sirupsen/logrus"
)

const (
	IndexAction  = "indices:data/write/index"
	GetAction    = "indices:data/read/get"
	DeleteAction = "indices:data/write/delete"
)

type indexRequest struct {
	Index   string
	Id      string
	Source  []byte
	ShardId state.ShardId
}

func (r *indexRequest) toBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func indexRequestFromBytes(b []byte) *indexRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var req indexRequest
	if err := decoder.Decode(&req); err != nil {
		logrus.Fatal(err)
	}
	return &req
}

type indexResponse struct {
	Err string
}

type RestIndexDoc struct {
	clusterService              *cluster.Service
	createIndexService          *cluster.MetadataCreateIndexService
	indicesService              *indices.Service
	indexNameExpressionResolver *indices.NameExpressionResolver
	transportService            *transport.Service
}

func NewRestIndexDoc(clusterService *cluster.Service, createIndexService *cluster.MetadataCreateIndexService, indicesService *indices.Service, indexNameExpressionResolver *indices.NameExpressionResolver, transportService *transport.Service) *RestIndexDoc {
	// Handle primary shard request
	transportService.RegisterRequestHandler(IndexAction, func(channel transport.ReplyChannel, req []byte) {
		logrus.Info("indexAction on primary shard")
		request := indexRequestFromBytes(req)

		indexService, _ := indicesService.IndexService(request.ShardId.Index.Uuid)
		indexShard, _ := indexService.Shard(request.ShardId.ShardId)

		var body map[string]interface{}
		// TODO :: json error handling
		if err := json.Unmarshal(request.Source, &body); err != nil {
			logrus.Fatal(err)
		}
		// TODO :: how to handle error on distributed system?
		if err := indexShard.Index(request.Id, body); err != nil {
			logrus.Fatal(err)
		}
		channel.SendMessage("", []byte("success"))
	})

	return &RestIndexDoc{
		clusterService:              clusterService,
		createIndexService:          createIndexService,
		indicesService:              indicesService,
		indexNameExpressionResolver: indexNameExpressionResolver,
		transportService:            transportService,
	}
}

func (h *RestIndexDoc) Handle(r *RestRequest, reply ResponseListener) {
	indexExpression := r.PathParams["index"]
	documentId := common.RandomBase64()
	var body map[string]interface{}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		logrus.Error(err)
		reply(RestResponse{
			StatusCode: 400,
			Body: map[string]interface{}{
				"err": err,
			},
		})
		return
	}

	clusterState := h.clusterService.State()
	indexName := h.indexNameExpressionResolver.ConcreteSingleIndex(*clusterState, indexExpression).Name
	if indexName == "" {
		req := cluster.CreateIndexClusterStateUpdateRequest{
			Index:    indexExpression,
			Mappings: []byte(`{ "properties": {} }`),
			Settings: map[string]interface{}{
				"number_of_shards": 1.0,
			},
		}
		// TODO :: refactor to callback function to execute after create
		// TODO :: index dynamic mapping
		h.createIndexService.CreateIndex(req)
		indexName = indexExpression
		clusterState = h.clusterService.State()

	}
	shardRouting := cluster.IndexShard(*clusterState, indexName, documentId).Primary
	indexRequest := indexRequest{
		Index:   indexName,
		Id:      documentId,
		Source:  r.Body,
		ShardId: shardRouting.ShardId,
	}
	h.transportService.SendRequest(*clusterState.Nodes.Nodes[shardRouting.CurrentNodeId], IndexAction, indexRequest.toBytes(), func(response []byte) {
		logrus.Info("callback success")
		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"_index":   indexName,
				"_type":    "_doc",
				"_id":      documentId,
				"_version": 1,
				"result":   "created",
				"_shards": map[string]interface{}{
					"total":      2,
					"successful": 1,
					"failed":     0,
				},
				"_seq_no":       0,
				"_primary_term": 1,
			},
		})
	})
}

type RestIndexDocId struct {
	clusterService              *cluster.Service
	createIndexService          *cluster.MetadataCreateIndexService
	indicesService              *indices.Service
	indexNameExpressionResolver *indices.NameExpressionResolver
	transportService            *transport.Service
}

func NewRestIndexDocId(clusterService *cluster.Service, createIndexService *cluster.MetadataCreateIndexService, indicesService *indices.Service, indexNameExpressionResolver *indices.NameExpressionResolver, transportService *transport.Service) *RestIndexDocId {
	// Handle primary shard request
	return &RestIndexDocId{
		clusterService:              clusterService,
		createIndexService:          createIndexService,
		indicesService:              indicesService,
		indexNameExpressionResolver: indexNameExpressionResolver,
		transportService:            transportService,
	}
}

func (h *RestIndexDocId) Handle(r *RestRequest, reply ResponseListener) {
	indexExpression := r.PathParams["index"]
	documentId := r.PathParams["id"]
	var body map[string]interface{}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		logrus.Error(err)
		reply(RestResponse{
			StatusCode: 400,
			Body: map[string]interface{}{
				"err": err,
			},
		})
		return
	}

	clusterState := h.clusterService.State()
	indexName := h.indexNameExpressionResolver.ConcreteSingleIndex(*clusterState, indexExpression).Name
	if indexName == "" {
		req := cluster.CreateIndexClusterStateUpdateRequest{
			Index:    indexExpression,
			Mappings: []byte(`{ "properties": {} }`),
			Settings: map[string]interface{}{
				"number_of_shards": 1.0,
			},
		}
		// TODO :: refactor to callback function to execute after create
		// TODO :: index dynamic mapping
		h.createIndexService.CreateIndex(req)
		indexName = indexExpression
		clusterState = h.clusterService.State()
	}
	shardRouting := cluster.IndexShard(*clusterState, indexExpression, documentId).Primary
	indexRequest := indexRequest{
		Index:   indexName,
		Id:      documentId,
		Source:  r.Body,
		ShardId: shardRouting.ShardId,
	}
	h.transportService.SendRequest(*clusterState.Nodes.Nodes[shardRouting.CurrentNodeId], IndexAction, indexRequest.toBytes(), func(response []byte) {
		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"_index":   indexName,
				"_type":    "_doc",
				"_id":      documentId,
				"_version": 1,
				"result":   "created",
				"_shards": map[string]interface{}{
					"total":      2,
					"successful": 1,
					"failed":     0,
				},
				"_seq_no":       0,
				"_primary_term": 1,
			},
		})
	})
}

type getRequest struct {
	Index   string
	Id      string
	ShardId state.ShardId
}

func (r *getRequest) toBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func getRequestFromBytes(b []byte) *getRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var req getRequest
	if err := decoder.Decode(&req); err != nil {
		logrus.Fatal(err)
	}
	return &req
}

type getResponse struct {
	Index   string
	Id      string
	ShardId state.ShardId
	Fields  map[string]interface{}
	Err     string
}

func (r *getResponse) toBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func getResponseFromBytes(b []byte) *getResponse {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var res getResponse
	if err := decoder.Decode(&res); err != nil {
		logrus.Fatal(err)
	}
	return &res
}

type RestGetDoc struct {
	clusterService              *cluster.Service
	indicesService              *indices.Service
	indexNameExpressionResolver *indices.NameExpressionResolver
	transportService            *transport.Service
}

func NewRestGetDoc(clusterService *cluster.Service, indicesService *indices.Service, indexNameExpressionResolver *indices.NameExpressionResolver, transportService *transport.Service) *RestGetDoc {
	transportService.RegisterRequestHandler(GetAction, func(channel transport.ReplyChannel, req []byte) {
		logrus.Info("getAction on shard")
		request := getRequestFromBytes(req)

		indexService, _ := indicesService.IndexService(request.ShardId.Index.Uuid)
		indexShard, _ := indexService.Shard(request.ShardId.ShardId)

		if doc, err := indexShard.Get(request.Id); err != nil {
			logrus.Warn(err)
			res := getResponse{
				Err: err.Error(),
			}
			channel.SendMessage("", res.toBytes())
		} else {
			res := getResponse{
				Index:   request.Index,
				Id:      request.Id,
				ShardId: request.ShardId,
				Fields:  doc,
			}
			channel.SendMessage("", res.toBytes())
		}
	})

	return &RestGetDoc{
		clusterService:              clusterService,
		indicesService:              indicesService,
		indexNameExpressionResolver: indexNameExpressionResolver,
		transportService:            transportService,
	}
}

func (h *RestGetDoc) Handle(r *RestRequest, reply ResponseListener) {
	indexExpression := r.PathParams["index"]
	documentId := r.PathParams["id"]

	clusterState := h.clusterService.State()
	indexName := h.indexNameExpressionResolver.ConcreteSingleIndex(*clusterState, indexExpression).Name
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
				Body: map[string]interface{}{
					"_index":        indexName,
					"_type":         "_doc",
					"_id":           documentId,
					"_version":      1,
					"_seq_no":       0,
					"_primary_term": 1,
					"found":         true,
					"_source":       res.Fields,
				},
			})
		}
	})

}

type deleteRequest struct {
	Index   string
	Id      string
	ShardId state.ShardId
}

func (r *deleteRequest) toBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func deleteRequestFromBytes(b []byte) *deleteRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var req deleteRequest
	if err := decoder.Decode(&req); err != nil {
		logrus.Fatal(err)
	}
	return &req
}

type deleteResponse struct {
	Err string
}

type RestDeleteDoc struct {
	clusterService              *cluster.Service
	indicesService              *indices.Service
	indexNameExpressionResolver *indices.NameExpressionResolver
	transportService            *transport.Service
}

func NewRestDeleteDoc(clusterService *cluster.Service, indicesService *indices.Service, indexNameExpressionResolver *indices.NameExpressionResolver, transportService *transport.Service) *RestDeleteDoc {
	transportService.RegisterRequestHandler(DeleteAction, func(channel transport.ReplyChannel, req []byte) {
		logrus.Info("deleteAction on primary shard")
		request := deleteRequestFromBytes(req)

		indexService, _ := indicesService.IndexService(request.ShardId.Index.Uuid)
		indexShard, _ := indexService.Shard(request.ShardId.ShardId)

		// TODO :: how to handle error on distributed system?
		if err := indexShard.Delete(request.Id); err != nil {
			logrus.Fatal(err)
		}

		channel.SendMessage("", []byte("success"))
	})

	return &RestDeleteDoc{
		clusterService:              clusterService,
		indicesService:              indicesService,
		indexNameExpressionResolver: indexNameExpressionResolver,
		transportService:            transportService,
	}
}

func (h *RestDeleteDoc) Handle(r *RestRequest, reply ResponseListener) {
	indexExpression := r.PathParams["index"]
	documentId := r.PathParams["id"]

	clusterState := h.clusterService.State()
	indexName := h.indexNameExpressionResolver.ConcreteSingleIndex(*clusterState, indexExpression).Name
	shardRouting := cluster.IndexShard(*clusterState, indexName, documentId).Primary
	deleteRequest := deleteRequest{
		Index:   indexName,
		Id:      documentId,
		ShardId: shardRouting.ShardId,
	}

	h.transportService.SendRequest(*clusterState.Nodes.Nodes[shardRouting.CurrentNodeId], DeleteAction, deleteRequest.toBytes(), func(response []byte) {
		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"_index":   indexName,
				"_type":    "_doc",
				"_id":      documentId,
				"_version": 2,
				"result":   "deleted",
				"_shards": map[string]interface{}{
					"total":      2,
					"successful": 1,
					"failed":     0,
				},
				"_seq_no":       1,
				"_primary_term": 1,
			},
		})
	})
}

type RestHeadDoc struct{}

func (h *RestHeadDoc) Handle(r *RestRequest, reply ResponseListener) {
	reply(RestResponse{
		StatusCode: 200,
	})
}
