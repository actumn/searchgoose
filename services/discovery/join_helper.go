package discovery

type JoinHelper struct {
}

type JoinAccumulator interface {
	handleJoinRequest(sender Node)
}

type InitialJoinAccumulator struct {
}

func (a *InitialJoinAccumulator) handleJoinRequest(sender Node) {

}

type CandidateJoinAccumulator struct {
}
