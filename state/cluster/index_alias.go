package cluster

import (
	"github.com/actumn/searchgoose/state"
	"github.com/sirupsen/logrus"
)

type AliasAction struct {
	Type  string
	Index string
	Alias string
}

type IndicesAliasesClusterStateUpdateRequest struct {
	Actions []AliasAction
}

type MetadataIndexAliasService struct {
	clusterService state.ClusterService
}

func NewMetadataIndexAliasService(clusterService state.ClusterService) *MetadataIndexAliasService {
	return &MetadataIndexAliasService{
		clusterService: clusterService,
	}
}

func (s *MetadataIndexAliasService) IndicesAliases(req IndicesAliasesClusterStateUpdateRequest) {
	s.clusterService.SubmitStateUpdateTask(func(current state.ClusterState) state.ClusterState {
		return s.applyAliasesAction(current, req.Actions)
	})
}

func (s *MetadataIndexAliasService) applyAliasesAction(current state.ClusterState, actions []AliasAction) state.ClusterState {
	for _, action := range actions {
		switch action.Type {
		case "add":
			current.Metadata.Indices[action.Index].Aliases[action.Alias] = state.AliasMetadata{
				Alias: action.Alias,
			}
		case "remove":
			logrus.Warn("Not yet implemented remove ", action.Index, " ", action.Alias)
		default:
			logrus.Fatal("Unknown action: ", action.Type)
		}
	}

	for _, indexMetadata := range current.Metadata.Indices {
		for _, aliasMetadata := range indexMetadata.Aliases {
			current.Metadata.IndicesLookup[indexMetadata.Index.Name] = state.IndexAbstractionAlias{
				AliasName:  aliasMetadata.Alias,
				WriteIndex: indexMetadata,
			}
		}
	}

	return current
}
