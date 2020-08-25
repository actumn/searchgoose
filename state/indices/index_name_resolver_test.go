package indices

import (
	"github.com/actumn/searchgoose/state"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNameExpressionResolver_concreteIndexNames(t *testing.T) {
	// Arrange
	resolver := NameExpressionResolver{}
	clusterState := state.ClusterState{
		Name: "_name",
		Metadata: state.Metadata{
			Indices: map[string]state.IndexMetadata{
				"foo": {
					Index: state.Index{
						Name: "foo",
					},
				},
				"foobar": {
					Index: state.Index{
						Name: "foobar",
					},
				},
				"foofoo-closed": {
					Index: state.Index{
						Name: "foofoo-closed",
					},
				},
				"foofoo": {
					Index: state.Index{
						Name: "foofoo",
					},
				},
			},
		},
	}

	// Action
	results := resolver.concreteIndexNames(clusterState, "foo")

	// Assert
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "foo", results[0])
}
