package persist

import (
	"github.com/actumn/searchgoose/services/metadata"
)

var (
	NoOnDiskState = OnDiskState{
		Metadata: metadata.EmptyMetadata,
	}
)

type OnDiskState struct {
	Id                  string
	DataPath            string
	CurrentTerm         int64
	LastAcceptedVersion int64
	Metadata            metadata.Metadata
}

func (s *OnDiskState) empty() bool {
	return s == &NoOnDiskState
}

type PersistedState interface {
	GetLastAcceptedState() *metadata.ClusterState
}
