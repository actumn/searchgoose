package actions

type RestIndexDoc struct{}

func (h *RestIndexDoc) Handle(r *RestRequest, reply ResponseListener) {
	indexName := r.PathParams["index"]

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

type RestIndexDocId struct{}

func (h *RestIndexDocId) Handle(r *RestRequest, reply ResponseListener) {
	indexName := r.PathParams["index"]
	idName := r.PathParams["id"]

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"_index":   indexName,
			"_type":    "_doc",
			"_id":      idName,
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

type RestGetDoc struct{}

func (h *RestGetDoc) Handle(r *RestRequest, reply ResponseListener) {
	indexName := r.PathParams["index"]
	idName := r.PathParams["id"]

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"_index":        indexName,
			"_type":         "_doc",
			"_id":           idName,
			"_version":      1,
			"_seq_no":       0,
			"_primary_term": 1,
			"found":         true,
			"_source": map[string]interface{}{
				"@timestamp": "2099-11-15T13:12:00",
				"message":    "GET /search HTTP/1.1 200 1070000",
				"user": map[string]interface{}{
					"id": "kimchy",
				},
			},
		},
	})
}

type RestHeadDoc struct{}

func (h *RestHeadDoc) Handle(r *RestRequest, reply ResponseListener) {
	reply(RestResponse{
		StatusCode: 200,
	})
}

type RestDeleteDoc struct{}

func (h *RestDeleteDoc) Handle(r *RestRequest, reply ResponseListener) {
	indexName := r.PathParams["index"]
	idName := r.PathParams["id"]

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"_index":   indexName,
			"_type":    "_doc",
			"_id":      idName,
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
