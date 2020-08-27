package actions

type RestIndexSearch struct{}

func (h *RestIndexSearch) Handle(r *RestRequest, reply ResponseListener) {
	indexName := r.PathParams["index"]

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"took":      0,
			"timed_out": false,
			"_shards": map[string]interface{}{
				"total":      1,
				"successful": 1,
				"skipped":    0,
				"failed":     0,
			},
			"hits": map[string]interface{}{
				"total": map[string]interface{}{
					"value":    1,
					"relation": "eq",
				},
				"max_score": 1.0,
				"hits": []interface{}{
					map[string]interface{}{
						"_index": indexName,
						"_type":  "_doc",
						"_id":    "1",
						"_score": 1.0,
						"_source": map[string]interface{}{
							"@timestamp": "2099-11-15T13:12:00",
							"message":    "GET /search HTTP/1.1 200 1070000",
							"user": map[string]interface{}{
								"id": "kimchy",
							},
						},
					},
				},
			},
		},
	})
}
