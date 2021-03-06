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

func NewGatewayMetaState() *GatewayMetaState {
	return &GatewayMetaState{}
}

func (m *GatewayMetaState) Start(
	transportService *transport.Service,
	clusterService *cluster.Service,
	persistedClusterStateService *persist.ClusterStateService) {
	onDiskState := persistedClusterStateService.LoadBestOnDiskState()

	clusterState := &state.ClusterState{
		Name: "searchgoose-testCluster",
		Nodes: &state.Nodes{
			Nodes: map[string]state.Node{
				transportService.LocalNode.Id: transportService.GetLocalNode(),
			},
			MasterNodes: map[string]state.Node{
				transportService.LocalNode.Id: transportService.GetLocalNode(),
			},
			DataNodes: map[string]state.Node{
				transportService.LocalNode.Id: transportService.GetLocalNode(),
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
	PersistedClusterStateService *persist.ClusterStateService
	CurrentTerm                  int64
	LastAcceptedState            *state.ClusterState
}

func (s *BlevePersistedState) GetLastAcceptedState() *state.ClusterState {
	return s.LastAcceptedState
}

func (s *BlevePersistedState) SetLastAcceptedState(state *state.ClusterState) {
	s.LastAcceptedState = state
}
