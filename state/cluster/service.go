package cluster

import "github.com/actumn/searchgoose/state"

type Service struct {
	//Settings       Settings
	ApplierService *ApplierService
	MasterService  *MasterService
}

func NewService() *Service {
	return &Service{
		ApplierService: newApplierService(),
		MasterService:  newMasterService(),
	}
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

func newApplierService() *ApplierService {
	return &ApplierService{}
}

func (s *ApplierService) OnNewState(clusterState *state.ClusterState) {
	// TODO:: goroutine 으로 구현하면 좋을 것 같다. (s.start() 해서)
	//changedEvent := state.ClusterChangedEvent{
	//	State:     *clusterState,
	//	PrevState: *s.ClusterState,
	//}

	s.ClusterState = clusterState

}

type MasterService struct {
	ClusterState        *state.ClusterState
	ClusterStatePublish func(event state.ClusterChangedEvent)
	// Publisher func
}

func newMasterService() *MasterService {
	return &MasterService{}
}

func (s *MasterService) submitStateUpdateTask(task state.ClusterStateUpdateTask) {
	// TODO:: goroutine 으로 구현하면 좋을 것 같다. (s.start() 해서)
	newState := task(*s.ClusterState)

	clusterChangedEvent := state.ClusterChangedEvent{
		State:     newState,
		PrevState: *s.ClusterState,
	}

	// publish ... 어떻게하지?
	s.ClusterStatePublish(clusterChangedEvent)
}
