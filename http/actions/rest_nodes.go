package actions

import (
	"github.com/actumn/searchgoose/state/cluster"
)

type RestNodes struct {
	clusterService *cluster.Service
}

func NewRestNodes(clusterService *cluster.Service) *RestNodes {
	return &RestNodes{
		clusterService: clusterService,
	}
}

func (h *RestNodes) Handle(r *RestRequest, reply ResponseListener) {
	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"nodes": map[string]interface{}{
				cluster.GenerateNodeId(): map[string]interface{}{
					"ip":      "127.0.0.1",
					"version": "7.8.0",
					"http": map[string]interface{}{
						"public_address": "127.0.0.1:8080",
					},
				},
			},
		},
	})
}
