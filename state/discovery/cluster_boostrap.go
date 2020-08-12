package discovery

type ClusterBootstrapService struct {
}

func (s *ClusterBootstrapService) ScheduleUnconfiguredBootstrap() {
	s.startBootstrap()
}

func (s *ClusterBootstrapService) startBootstrap() {

}
