package actions

import "github.com/actumn/searchgoose/state/cluster"

type RestRefresh struct {
	clusterService *cluster.Service
}

func NewRestRefresh(clusterService *cluster.Service) *RestRefresh {
	return &RestRefresh{
		clusterService: clusterService,
	}
}

func (h *RestRefresh) Handle(r *RestRequest, reply ResponseListener) {
	// TODO:: broadcast request and reply appropriate response
	index := r.PathParams["index"]
	clusterState := h.clusterService.State()
	shardCount := clusterState.Metadata.Indices[index].NumberOfShards
	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"_shard": map[string]interface{}{
				"total":      shardCount,
				"successful": shardCount,
				"failed":     0,
			},
		},
	})
}
