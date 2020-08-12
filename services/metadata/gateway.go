package metadata

import (
	"github.com/actumn/searchgoose/services"
	"github.com/actumn/searchgoose/services/cluster"
	"github.com/actumn/searchgoose/services/persist"
	"github.com/actumn/searchgoose/services/transport"
)

func prepareInitialClusterState() {

}

type GatewayMetaState struct {
	PersistedState services.PersistedState
}

func (m *GatewayMetaState) Start(
	transportService transport.Service,
	clusterService cluster.Service,
	persistedClusterStateService persist.ClusterStateService) {
	onDiskState := persistedClusterStateService.LoadBestOnDiskState()

	clusterState := &services.ClusterState{
		Name: "searchgoose-testCluster",
		Nodes: &services.Nodes{
			Nodes: map[string]*services.Node{
				transportService.LocalNode.Id: transportService.LocalNode,
			},
			LocalNodeId: transportService.LocalNode.Id,
		},
		Version:  onDiskState.LastAcceptedVersion,
		Metadata: onDiskState.Metadata,
	}

	m.PersistedState = &BlevePersistedState{
		PersistedClusterStateService: persistedClusterStateService,
		CurrentTerm:                  onDiskState.CurrentTerm,
		LastAcceptedState:            clusterState,
	}
}

type BlevePersistedState struct {
	PersistedClusterStateService persist.ClusterStateService
	CurrentTerm                  int64
	LastAcceptedState            *services.ClusterState
}

func (s *BlevePersistedState) GetLastAcceptedState() *services.ClusterState {
	return s.LastAcceptedState
}
