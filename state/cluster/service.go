package cluster

import "github.com/actumn/searchgoose/state"

type Service struct {
	Settings       Settings
	ApplierService ApplierService
	//MasterService
}

func (s *Service) State() state.ClusterState {
	return s.ApplierService.ClusterState
}

type ApplierService struct {
	ClusterState state.ClusterState
}
