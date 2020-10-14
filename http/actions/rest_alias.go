package actions

import (
	"encoding/json"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/sirupsen/logrus"
)

type RestGetIndexAlias struct {
}

func NewRestGetIndexAlias() *RestGetIndexAlias {
	return &RestGetIndexAlias{}
}

func (h *RestGetIndexAlias) Handle(r *RestRequest, reply ResponseListener) {
	name := r.PathParams["name"]

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			name: map[string]interface{}{},
		},
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
