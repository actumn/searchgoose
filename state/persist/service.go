package persist

import "github.com/actumn/searchgoose/state"

type ClusterStateService struct {
}

func (c *ClusterStateService) LoadBestOnDiskState() *state.OnDiskState {
	return &state.NoOnDiskState
}
