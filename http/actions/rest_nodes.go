package actions

import (
	"github.com/actumn/searchgoose/services/cluster"
)

type RestNodes struct{}

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
