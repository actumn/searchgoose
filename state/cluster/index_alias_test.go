package cluster

import (
	"github.com/actumn/searchgoose/state"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestMetadataIndexAliasService_applyAliasesAction(t *testing.T) {
	// Arrange
	service := MetadataIndexAliasService{}
	currentState := state.ClusterState{}
	actions := []AliasAction{
		{
			Type:  "remove",
			Index: ".kibana_task_manager",
			Alias: ".kibana_task_manager",
		},
		{
			Type:  "add",
			Index: ".kibana_task_manager_1",
			Alias: ".kibana_task_manager",
		},
	}
	// Action
	result := service.applyAliasesAction(currentState, actions)

	// Assert
	logrus.Info(result)
}
