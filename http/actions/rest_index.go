package actions

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
}

type RestPutIndex struct{}

func (h *RestPutIndex) Handle(r *RestRequest, reply ResponseListener) {
	indexName := r.PathParams["index"]

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"acknowledged":        true,
			"shards_acknowledged": true,
			"index":               indexName,
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
