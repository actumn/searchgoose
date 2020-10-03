package actions

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/actumn/searchgoose/common"
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/actumn/searchgoose/state/transport"
	"log"
)

const (
	IndexAction = "indices_write"
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
		log.Fatalln(err)
	}
	return buffer.Bytes()
}

func indexRequestFromBytes(b []byte) *indexRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var req indexRequest
	if err := decoder.Decode(&req); err != nil {
		log.Fatalln(err)
	}
	return &req
}

type indexResponse struct {
}

type RestIndexDoc struct {
	ClusterService   *cluster.Service
	IndicesService   *indices.Service
	TransportService *transport.Service
}

func NewRestIndexDoc(clusterService *cluster.Service, indicesService *indices.Service, transportService *transport.Service) *RestIndexDoc {
	// Handle primary shard request
	transportService.RegisterRequestHandler(IndexAction, func(channel transport.ReplyChannel, req []byte) []byte {
		log.Println("indexAction on primary shard")
		request := indexRequestFromBytes(req)

		clusterState := clusterService.State()
		index := clusterState.Metadata.Indices[request.Index].Index
		uuid := index.Uuid

		indexService, _ := indicesService.IndexService(uuid)
		indexShard, _ := indexService.Shard(request.ShardId.ShardId)

		var body map[string]interface{}
		// TODO :: json error handling
		if err := json.Unmarshal(request.Source, &body); err != nil {
			log.Fatalln(err)
		}
		// TODO :: how to handle error on distributed system?
		if err := indexShard.Index(request.Id, body); err != nil {
			log.Fatalln(err)
		}
		return []byte("success")
	})

	return &RestIndexDoc{
		ClusterService:   clusterService,
		IndicesService:   indicesService,
		TransportService: transportService,
	}
}

func (h *RestIndexDoc) Handle(r *RestRequest, reply ResponseListener) {
	indexName := r.PathParams["index"]
	documentId := common.RandomBase64()
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

	// TODO :: auto create index if absent.
	clusterState := h.ClusterService.State()
	shardRouting := cluster.IndexShard(*clusterState, indexName, documentId).Primary
	indexRequest := indexRequest{
		Index:   indexName,
		Id:      documentId,
		Source:  r.Body,
		ShardId: shardRouting.ShardId,
	}
	h.TransportService.SendRequest(*clusterState.Nodes.Nodes[shardRouting.CurrentNodeId], IndexAction, indexRequest.toBytes(), func(response []byte) {
		log.Println("callback success")
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
	ClusterService   *cluster.Service
	IndicesService   *indices.Service
	TransportService *transport.Service
}

func (h *RestIndexDocId) Handle(r *RestRequest, reply ResponseListener) {
	indexName := r.PathParams["index"]
	documentId := r.PathParams["id"]
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

	// TODO :: auto create index if absent.
	clusterState := h.ClusterService.State()
	shardRouting := cluster.IndexShard(*clusterState, indexName, documentId).Primary
	indexRequest := indexRequest{
		Index:   indexName,
		Id:      documentId,
		Source:  r.Body,
		ShardId: shardRouting.ShardId,
	}
	h.TransportService.SendRequest(*clusterState.Nodes.Nodes[shardRouting.CurrentNodeId], IndexAction, indexRequest.toBytes(), func(response []byte) {
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

type RestGetDoc struct {
	ClusterService *cluster.Service
	IndicesService *indices.Service
}

func (h *RestGetDoc) Handle(r *RestRequest, reply ResponseListener) {
	indexName := r.PathParams["index"]
	documentId := r.PathParams["id"]

	// TODO :: document routing on primary shard with RoutingTable
	clusterState := h.ClusterService.State()
	index := clusterState.Metadata.Indices[indexName].Index
	uuid := index.Uuid

	indexService, _ := h.IndicesService.IndexService(uuid)
	//shard := cluster.GetShards(*clusterState, index.Name, documentId).Primary
	//indexShard, _ := indexService.Shard(shard.ShardId.ShardId)
	indexShard, _ := indexService.Shard(0)
	if doc, err := indexShard.Get(documentId); err != nil {
		reply(RestResponse{
			StatusCode: 400,
			Body: map[string]interface{}{
				"err": err,
			},
		})
		return
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
				"_source":       doc,
			},
		})
	}
}

type RestDeleteDoc struct {
	ClusterService *cluster.Service
	IndicesService *indices.Service
}

func (h *RestDeleteDoc) Handle(r *RestRequest, reply ResponseListener) {
	indexName := r.PathParams["index"]
	documentId := r.PathParams["id"]

	// TODO :: document routing on primary shard with RoutingTable
	clusterState := h.ClusterService.State()
	index := clusterState.Metadata.Indices[indexName].Index
	uuid := index.Uuid

	indexService, _ := h.IndicesService.IndexService(uuid)
	//shard := cluster.GetShards(*clusterState, index.Name, documentId).Primary
	//indexShard, _ := indexService.Shard(shard.ShardId.ShardId)
	indexShard, _ := indexService.Shard(0)
	if err := indexShard.Delete(documentId); err != nil {
		log.Fatalln(err)
	}

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
}

type RestHeadDoc struct{}

func (h *RestHeadDoc) Handle(r *RestRequest, reply ResponseListener) {
	reply(RestResponse{
		StatusCode: 200,
	})
}
