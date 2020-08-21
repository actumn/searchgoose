package state

var (
	EmptyMetadata = Metadata{}
)

type IndexMetadataState int

const (
	OPEN IndexMetadataState = iota
	CLOSE
)

type Metadata struct {
	ClusterUUID string
	Version     int64
	//Coordination CoordinationMetadata
	Indices map[string]IndexMetadata
	//Templates    map[string]IndexTemplateMetadata
}

type Index struct {
	Name string
	Uuid string
}

type IndexMetadata struct {
	Index              Index
	RoutingNumShards   int
	RoutingNumReplicas int
	Version            int64
	State              IndexMetadataState
	Mapping            map[string]MappingMetadata
	//Settings Settings
}

type MappingMetadata struct {
	Type   string
	Source []byte
}

//type IndexTemplateMetadata struct {
//}
//
//type CoordinationMetadata struct {
//}
