package actions

import (
	"encoding/json"
	"fmt"
)

type RestGetIndices struct{}

func (h *RestGetIndices) Handle(r *RestRequest, reply ResponseListener) {
	indexName := r.PathParams["index"]

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"error": map[string]interface{}{
				"root_cause": []map[string]interface{}{
					{
						"type":          "index_not_found_exception",
						"reason":        "no such index [" + indexName + "]",
						"resource.type": "index_or_alias",
						"resource.id":   ".kibana",
						"index_uuid":    "_na_",
						"index":         ".kibana",
					},
				},
				"type":          "index_not_found_exception",
				"reason":        "no such index [" + indexName + "]",
				"resource.type": "index_or_alias",
				"resource.id":   indexName,
				"index_uuid":    "_na_",
				"index":         indexName,
			},
			"status": 404,
		},
	})
}

type RestPutIndices struct{}

func (h *RestPutIndices) Handle(r *RestRequest, reply ResponseListener) {
	index := r.PathParams["index"]

	var body map[string]interface{}
	err := json.Unmarshal(r.Body, &body)
	if err != nil {
		reply(RestResponse{
			StatusCode: 400,
			Body: map[string]interface{}{
				"err": err,
			},
		})
	}
	fmt.Print(body)

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"acknowledged":        true,
			"shards_acknowledged": true,
			"index":               index,
		},
	})
}
