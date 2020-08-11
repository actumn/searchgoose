package persist

type ClusterStateService struct {
}

func (c *ClusterStateService) LoadBestOnDiskState() *OnDiskState {
	return &NoOnDiskState
}
