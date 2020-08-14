package state

type Metadata struct {
	ClusterUUID  string
	Version      int64
	Coordination CoordinationMetadata
	Indices      map[string]IndexMetadata
	Templates    map[string]IndexTemplateMetadata
}

var (
	EmptyMetadata = Metadata{}
)

type IndexMetadata struct {
}

type IndexTemplateMetadata struct {
}

type CoordinationMetadata struct {
}
