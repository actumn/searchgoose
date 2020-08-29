package actions

import (
	"encoding/json"
	"github.com/actumn/searchgoose/common"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"log"
)

type RestIndexDoc struct {
	ClusterService *cluster.Service
	IndicesService *indices.Service
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
	// TODO :: document routing on primary shard with RoutingTable
	clusterState := h.ClusterService.State()
	index := clusterState.Metadata.Indices[indexName].Index
	uuid := index.Uuid

	indexService, _ := h.IndicesService.IndexService(uuid)
	indexShard, _ := indexService.Shard(0)
	if err := indexShard.Index(documentId, body); err != nil {
		log.Fatalln(err)
	}

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"_index":   indexName,
			"_type":    "_doc",
			"_id":      "1", //random id
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
}

type RestIndexDocId struct {
	ClusterService *cluster.Service
	IndicesService *indices.Service
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
	// TODO :: document routing on primary shard with RoutingTable
	clusterState := h.ClusterService.State()
	index := clusterState.Metadata.Indices[indexName].Index
	uuid := index.Uuid

	indexService, _ := h.IndicesService.IndexService(uuid)
	indexShard, _ := indexService.Shard(0)
	if err := indexShard.Index(documentId, body); err != nil {
		log.Fatalln(err)
	}

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

type RestHeadDoc struct{}

func (h *RestHeadDoc) Handle(r *RestRequest, reply ResponseListener) {
	reply(RestResponse{
		StatusCode: 200,
	})
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
