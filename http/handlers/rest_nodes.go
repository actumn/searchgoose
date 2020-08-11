package handlers

import "github.com/actumn/searchgoose/services"

type RestNodes struct{}

func (h *RestNodes) Handle(r *RestRequest) (interface{}, error) {
	return map[string]interface{}{
		"nodes": map[string]interface{}{
			services.GenerateNodeId(): map[string]interface{}{
				"ip":      "127.0.0.1",
				"version": "7.8.0",
				"http": map[string]interface{}{
					"public_address": "127.0.0.1:8080",
				},
			},
		},
	}, nil
}
