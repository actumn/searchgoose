package actions

import "github.com/actumn/searchgoose/state/cluster"

type RestMain struct {
	clusterService *cluster.Service
}

func NewRestMain(clusterService *cluster.Service) *RestMain {
	return &RestMain{
		clusterService: clusterService,
	}
}

func (h *RestMain) Handle(r *RestRequest, reply ResponseListener) {
	clusterState := h.clusterService.State()
	localNodeId := clusterState.Nodes.LocalNodeId
	localNode := clusterState.Nodes.Nodes[localNodeId]

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"name":         localNode.Name,
			"cluster_name": clusterState.Name,
			"cluster_uuid": clusterState.StateUUID,
			"version": map[string]interface{}{
				"number": "0.0.0",
			},
		},
	})
}
