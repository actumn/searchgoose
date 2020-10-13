package discovery

import (
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/transport"
)

type ClusterBootstrapService struct {
	TransportService *transport.Service
}

func NewClusterBootstrapService(service *transport.Service) *ClusterBootstrapService {
	return &ClusterBootstrapService{
		TransportService: service,
	}
}

func (s *ClusterBootstrapService) onFoundPeersUpdated() {
	s.startBootstrap()
}

/*
func (s *ClusterBootstrapService) ScheduleUnconfiguredBootstrap() {
	s.startBootstrap()
}
*/

func (s *ClusterBootstrapService) doBootstrap(configuration state.VotingConfiguration) {

}

func (s *ClusterBootstrapService) startBootstrap() {
}
