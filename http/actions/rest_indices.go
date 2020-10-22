package actions

import (
	"github.com/actumn/searchgoose/state/indices"
)

type RestIndicesStatsAction struct {
	indicesService *indices.Service
}

func NewRestIndicesStatsAction(indicesService *indices.Service) *RestIndicesStatsAction {
	return &RestIndicesStatsAction{
		indicesService: indicesService,
	}
}

func (h *RestIndicesStatsAction) Handle(r *RestRequest, reply ResponseListener) {
	// TODO:: reply based on indices service and all cluster shards information
	indicesMap := map[string]interface{}{
		".kibana_task_manager_2": map[string]interface{}{
			"uuid": "3qXOKMs-QYS0RL2ErQFubQ",
		},
	}
	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"_shards": map[string]interface{}{
				"total":      1,
				"successful": 1,
				"failed":     0,
			},
			"_all": map[string]interface{}{
				"primaries": map[string]interface{}{},
				"total":     map[string]interface{}{},
			},
			"indices": indicesMap,
		},
	})
}
