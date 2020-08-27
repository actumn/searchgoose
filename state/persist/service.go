package persist

import "github.com/actumn/searchgoose/state"

type ClusterStateService struct {
}

func NewClusterStateService() *ClusterStateService {
	return &ClusterStateService{}
}

func (c *ClusterStateService) LoadBestOnDiskState() *state.OnDiskState {
	return &state.NoOnDiskState
}
