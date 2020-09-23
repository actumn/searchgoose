package actions

import (
	"encoding/json"
	"github.com/actumn/searchgoose/state/cluster"
)

type RestGetIndex struct{}

func (h *RestGetIndex) Handle(r *RestRequest, reply ResponseListener) {
	indexName := r.PathParams["index"]

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			indexName: map[string]interface{}{
				"aliases":  map[string]interface{}{},
				"mappings": map[string]interface{}{},
				"settings": map[string]interface{}{
					"index": map[string]interface{}{
						"creation_date":      "1597382566866",
						"number_of_shards":   "1",
						"number_of_replicas": "1",
						"uuid":               "8LfuNgQZRrCODwai_Vfcug",
						"version": map[string]interface{}{
							"created": "7080299",
						},
						"provided_name": indexName,
					},
				},
			},
		},
	})

	//reply(RestResponse{
	//	StatusCode: 200,
	//	Body: map[string]interface{}{
	//		"error": map[string]interface{}{
	//			"root_cause": []map[string]interface{}{
	//				{
	//					"type":          "index_not_found_exception",
	//					"reason":        "no such index [" + indexName + "]",
	//					"resource.type": "index_or_alias",
	//					"resource.id":   ".kibana",
	//					"index_uuid":    "_na_",
	//					"index":         ".kibana",
	//				},
	//			},
	//			"type":          "index_not_found_exception",
	//			"reason":        "no such index [" + indexName + "]",
	//			"resource.type": "index_or_alias",
	//			"resource.id":   indexName,
	//			"index_uuid":    "_na_",
	//			"index":         indexName,
	//		},
	//		"status": 404,
	//	},
	//})
}

type RestPutIndex struct {
	CreateIndexService *cluster.MetadataCreateIndexService
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

type RestDeleteIndex struct{}

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
