package actions

import (
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/actumn/searchgoose/state/transport"
)

type RestCatTemplates struct{}

func (h *RestCatTemplates) Handle(r *RestRequest, reply ResponseListener) {
	reply(RestResponse{
		StatusCode: 200,
		Body:       []interface{}{},
	})
}

type RestCatNodes struct {
	clusterService   *cluster.Service
	transportService *transport.Service
}

func NewRestCatNodes(clusterService *cluster.Service, transportService *transport.Service) *RestCatNodes {
	return &RestCatNodes{
		clusterService:   clusterService,
		transportService: transportService,
	}
}

func (h *RestCatNodes) Handle(r *RestRequest, reply ResponseListener) {
	// TODO :: resolve nodes list from cluster state and broadcasting
	clusterState := h.clusterService.State()
	var nodesList []map[string]interface{}
	for n, node := range clusterState.Nodes.Nodes {
		nodesList = append(nodesList, map[string]interface{}{
			"id":         node.Id,
			"m":          "*", // master
			"n":          n + node.Name,
			"u":          "44m",     // uptime
			"role":       "dilmrt",  // monitor.role
			"hc":         "156.8mb", // heap current
			"hm":         "512mb",   // heap max
			"hp":         "30",      // heap percent
			"ip":         node.HostAddress,
			"dt":         "468.4gb", // disk total
			"du":         "267.4gb", // disk used
			"disk.avail": "200.9gb", // disk available
			"l":          "-1",      //
		})
	}
	reply(RestResponse{
		StatusCode: 200,
		Body:       nodesList,
	})
}

type RestCatIndices struct {
	clusterService              *cluster.Service
	indexNameExpressionResolver *indices.NameExpressionResolver
	transportService            *transport.Service
}

func NewRestCatIndices(clusterService *cluster.Service, indexNameExpressionResolver *indices.NameExpressionResolver, transportService *transport.Service) *RestCatIndices {
	return &RestCatIndices{
		clusterService:              clusterService,
		indexNameExpressionResolver: indexNameExpressionResolver,
		transportService:            transportService,
	}
}

func (h *RestCatIndices) Handle(r *RestRequest, reply ResponseListener) {
	clusterState := h.clusterService.State()
	indicesList := []map[string]interface{}{}
	for _, indexMetadata := range clusterState.Metadata.Indices {
		// TODO :: resolve indices information from broadcasting
		indicesList = append(indicesList, map[string]interface{}{
			"health": "green",
			"status": "open",
			"index":  indexMetadata.Index.Name,
			"uuid":   indexMetadata.Index.Uuid,
			//"pri":            "1",
			//"rep":            "0",
			//"docs.count":     "100",
			//"docs.deleted":   "0",
			//"store.size":     "208b",
			//"pri.store.size": "208b",
		})
	}

	reply(RestResponse{
		StatusCode: 200,
		Body:       indicesList,
	})
}
