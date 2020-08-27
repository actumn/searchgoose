package state

type ClusterService interface {
	State() *ClusterState
	SubmitStateUpdateTask(task ClusterStateUpdateTask)
}

type ClusterStateUpdateTask func(s ClusterState) ClusterState

type ClusterChangedEvent struct {
	State     ClusterState
	PrevState ClusterState
}
