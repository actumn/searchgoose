package actions

import (
	"encoding/json"
	"fmt"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/sirupsen/logrus"
	"strings"
)

type RestGetIndexAlias struct {
	clusterService              *cluster.Service
	indexNameExpressionResolver *indices.NameExpressionResolver
}

func NewRestGetIndexAlias(clusterService *cluster.Service, indexNameExpressionResolver *indices.NameExpressionResolver) *RestGetIndexAlias {
	return &RestGetIndexAlias{
		clusterService:              clusterService,
		indexNameExpressionResolver: indexNameExpressionResolver,
	}
}

func (h *RestGetIndexAlias) Handle(r *RestRequest, reply ResponseListener) {
	clusterState := h.clusterService.State()
	aliasesExpressions := strings.Split(r.PathParams["name"], ",")
	concreteIndices := h.indexNameExpressionResolver.ConcreteIndexNames(*clusterState, "*")
	aliasesMap := clusterState.Metadata.FindAliases(aliasesExpressions, concreteIndices)

	if len(aliasesMap) == 0 {
		logrus.Warn("alias missing")
		reply(RestResponse{
			StatusCode: 404,
			Body: map[string]interface{}{
				"error":  fmt.Sprintf("alias [.kibana] missing"),
				"status": 404,
			},
		})
		return
	}

	response := map[string]map[string]interface{}{}
	for idx, aliases := range aliasesMap {
		response[idx] = map[string]interface{}{
			"aliases": map[string]interface{}{},
		}

		for _, alias := range aliases {
			response[idx]["aliases"] = map[string]interface{}{
				alias.Alias: map[string]interface{}{},
			}
		}
	}

	reply(RestResponse{
		StatusCode: 200,
		Body:       response,
	})
}

type RestPostIndexAlias struct {
	indexAliasesService *cluster.MetadataIndexAliasService
}

func NewRestPostIndexAlias(indexAliasesService *cluster.MetadataIndexAliasService) *RestPostIndexAlias {
	return &RestPostIndexAlias{
		indexAliasesService: indexAliasesService,
	}
}

func (h *RestPostIndexAlias) Handle(r *RestRequest, reply ResponseListener) {
	var body map[string]interface{}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		logrus.Error(err)
		reply(RestResponse{
			StatusCode: 400,
			Body: map[string]interface{}{
				"err": err,
			},
		})
		return
	}

	actions := body["actions"].([]interface{})

	req := cluster.IndicesAliasesClusterStateUpdateRequest{
		Actions: []cluster.AliasAction{},
	}
	for _, action := range actions {
		action := action.(map[string]interface{})
		if add, existing := action["add"]; existing {
			add := add.(map[string]interface{})
			req.Actions = append(req.Actions, cluster.AliasAction{
				Type:  "add",
				Index: add["index"].(string),
				Alias: add["alias"].(string),
			})
		} else if remove, existing := action["remove"]; existing {
			remove := remove.(map[string]interface{})
			req.Actions = append(req.Actions, cluster.AliasAction{
				Type:  "remove",
				Index: remove["index"].(string),
				Alias: remove["alias"].(string),
			})
		}
	}

	h.indexAliasesService.IndicesAliases(req)

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"acknowledged": true,
		},
	})
}
