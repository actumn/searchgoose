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

func (s *Service) SubmitStateUpdateTask() {

}

type ApplierService struct {
	ClusterState *state.ClusterState
}

type MasterService struct {
}
