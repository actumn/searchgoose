package state

var (
	NoOnDiskState = OnDiskState{
		Metadata: EmptyMetadata,
	}
)

type OnDiskState struct {
	Id                  string
	DataPath            string
	CurrentTerm         int64
	LastAcceptedVersion int64
	Metadata            Metadata
}

func (s *OnDiskState) empty() bool {
	return s == &NoOnDiskState
}

type PersistedState interface {
	GetLastAcceptedState() *ClusterState
	SetLastAcceptedState(state *ClusterState)
}
