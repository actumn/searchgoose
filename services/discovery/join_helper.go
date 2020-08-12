package discovery

import "github.com/actumn/searchgoose/services"

type JoinHelper struct {
}

type JoinAccumulator interface {
	handleJoinRequest(sender services.Node)
}

type InitialJoinAccumulator struct {
}

func (a *InitialJoinAccumulator) handleJoinRequest(sender services.Node) {

}

type CandidateJoinAccumulator struct {
}
