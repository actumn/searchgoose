package metadata

import (
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/persist"
	"github.com/actumn/searchgoose/state/transport"
)

func prepareInitialClusterState() {

}

type GatewayMetaState struct {
	PersistedState state.PersistedState
}

func (m *GatewayMetaState) Start(
	transportService transport.Service,
	clusterService cluster.Service,
	persistedClusterStateService persist.ClusterStateService) {
	onDiskState := persistedClusterStateService.LoadBestOnDiskState()

	clusterState := &state.ClusterState{
		Name: "searchgoose-testCluster",
		Nodes: &state.Nodes{
			Nodes: map[string]*state.Node{
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
	LastAcceptedState            *state.ClusterState
}

func (s *BlevePersistedState) GetLastAcceptedState() *state.ClusterState {
	return s.LastAcceptedState
}
