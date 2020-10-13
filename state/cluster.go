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

func (e *ClusterChangedEvent) IndicesDeleted() []Index {
	var deleted []Index

	for _, index := range e.PrevState.Metadata.Indices {
		if _, existing := e.State.Metadata.Indices[index.Index.Name]; !existing {
			deleted = append(deleted, index.Index)
		}
	}

	return deleted
}
