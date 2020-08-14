package discovery

import "github.com/actumn/searchgoose/state"

type JoinHelper struct {
}

type JoinAccumulator interface {
	handleJoinRequest(sender state.Node)
}

type InitialJoinAccumulator struct {
}

func (a *InitialJoinAccumulator) handleJoinRequest(sender state.Node) {

}

type CandidateJoinAccumulator struct {
}
