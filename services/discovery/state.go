package discovery

import "github.com/actumn/searchgoose/services/persist"

type CoordinationState struct {
	LocalNode      *Node
	PersistedState persist.PersistedState
}
