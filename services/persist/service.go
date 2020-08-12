package persist

import "github.com/actumn/searchgoose/services"

type ClusterStateService struct {
}

func (c *ClusterStateService) LoadBestOnDiskState() *services.OnDiskState {
	return &services.NoOnDiskState
}
