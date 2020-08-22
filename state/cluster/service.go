package cluster

import "github.com/actumn/searchgoose/state"

type Service struct {
	Settings       Settings
	ApplierService ApplierService
	MasterService  MasterService
}

func (s *Service) State() *state.ClusterState {
	return s.ApplierService.ClusterState
}

func (s *Service) SubmitStateUpdateTask(task state.ClusterStateUpdateTask) {
	s.MasterService.submitStateUpdateTask(task)
}

type ApplierService struct {
	ClusterState *state.ClusterState
}

type MasterService struct {
	ClusterState *state.ClusterState
}

func (s *MasterService) submitStateUpdateTask(task state.ClusterStateUpdateTask) {
	// goroutine 으로 구현하면 좋을 것 같다.
	newState := task(*s.ClusterState)

	clusterChangedEvent := state.ClusterChangedEvent{
		State:     newState,
		PrevState: *s.ClusterState,
	}

	// publish ... 어떻게하지?

}
