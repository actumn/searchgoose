package indices

import (
	"github.com/actumn/searchgoose/state"
	"github.com/stretchr/testify/assert"
	"sort"
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
	results := resolver.ConcreteIndexNames(clusterState, "foo")

	// Assert
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "foo", results[0])

	results = resolver.ConcreteIndexNames(clusterState, "foobar")

	assert.Equal(t, 1, len(results))
	assert.Equal(t, "foobar", results[0])

	results = resolver.ConcreteIndexNames(clusterState, "foo*")

	sorted := []string{"foo", "foobar", "foofoo-closed", "foofoo"}
	sort.Strings(sorted)
	sort.Strings(results)
	assert.Equal(t, 4, len(results))
	assert.Equal(t, sorted, results)

	results = resolver.ConcreteIndexNames(clusterState, "foofoo*")

	sorted = []string{"foofoo", "foofoo-closed"}
	sort.Strings(sorted)
	sort.Strings(results)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, sorted, results)

	results = resolver.ConcreteIndexNames(clusterState, "bar")

	assert.Equal(t, []string(nil), results)
}
