package actions

import "github.com/actumn/searchgoose/state/indices"

type RestCatTemplates struct{}

func (h *RestCatTemplates) Handle(r *RestRequest, reply ResponseListener) {
	reply(RestResponse{
		StatusCode: 200,
		Body:       []interface{}{},
	})
}

type RestCatNodes struct {
}

func NewRestCatNodes() *RestCatNodes {
	return &RestCatNodes{}
}

func (h *RestCatNodes) Handle(r *RestRequest, reply ResponseListener) {
	// TODO :: resolve nodes list from cluster state and broadcasting
	reply(RestResponse{
		StatusCode: 200,
		Body: []interface{}{
			map[string]interface{}{
				"id":         "92_F",
				"m":          "*",
				"n":          "es-main",
				"u":          "44m",
				"role":       "dilmrt",
				"hc":         "156.8mb",
				"hm":         "512mb",
				"hp":         "30",
				"ip":         "172.28.0.1",
				"dt":         "468.4gb",
				"du":         "267.4gb",
				"disk.avail": "200.9gb",
				"l":          "2.62",
			},
			map[string]interface{}{
				"id":         "92_G",
				"m":          "*",
				"n":          "es-main2",
				"u":          "44m",
				"role":       "dilmrt",
				"hc":         "156.8mb",
				"hm":         "512mb",
				"hp":         "30",
				"ip":         "172.28.0.2",
				"dt":         "468.4gb",
				"du":         "267.4gb",
				"disk.avail": "200.9gb",
				"l":          "2.62",
			},
		},
	})
}

type RestCatIndices struct {
	indexNameExpressionResolver *indices.NameExpressionResolver
}

func NewRestCatIndices(indexNameExpressionResolver *indices.NameExpressionResolver) *RestCatIndices {
	return &RestCatIndices{
		indexNameExpressionResolver: indexNameExpressionResolver,
	}
}

func (h *RestCatIndices) Handle(r *RestRequest, reply ResponseListener) {
	// TODO :: resolve indices list from cluster state and broadcasting
	reply(RestResponse{
		StatusCode: 200,
		Body: []interface{}{
			map[string]interface{}{
				"health":         "green",
				"status":         "open",
				"index":          ".kibana_task_manager_2",
				"uuid":           "3qXOKMs-QYS0RL2ErQFubQ",
				"pri":            "1",
				"rep":            "0",
				"docs.count":     "0",
				"docs.deleted":   "0",
				"store.size":     "208b",
				"pri.store.size": "208b",
			},
			map[string]interface{}{
				"health":         "yellow",
				"status":         "open",
				"index":          ".elastichq",
				"uuid":           "HaeD3pq9TvyaSDPmFINvUQ",
				"pri":            "1",
				"rep":            "1",
				"docs.count":     "1",
				"docs.deleted":   "0",
				"store.size":     "6.4kb",
				"pri.store.size": "6.4kb",
			},
		},
	})
}
