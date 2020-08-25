package indices

import "github.com/actumn/searchgoose/state"

type NameExpressionResolver struct{}

func (r *NameExpressionResolver) concreteIndexNames(clusterState state.ClusterState, expression string) []string {
	return nil
}

func (r *NameExpressionResolver) concreteIndices(clusterState state.ClusterState, expression string) []state.Index {
	return nil
}
