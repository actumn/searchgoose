package actions

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

type RestPostIndices struct{}

//func (h *RestPostIndices) Handle(r *RestRequest) (interface{}, error) {
//	return map[string]interface{}{
//
//	}, nil
//}
