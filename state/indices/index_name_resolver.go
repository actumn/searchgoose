package indices

import (
	"github.com/actumn/searchgoose/state"
	"strings"
)

type NameExpressionResolver struct{}

func NewNameExpressionResolver() *NameExpressionResolver {
	return &NameExpressionResolver{}
}

//func (r *NameExpressionResolver) jsConcreteIndicesNames(clusterState state.ClusterState, expressions ...string) []string {
//
//}

func (r *NameExpressionResolver) ConcreteIndexNames(clusterState state.ClusterState, expression string) []string {
	var indexResult []string
	indiceResult := r.ConcreteIndices(clusterState, expression)

	if strings.HasSuffix(expression, "*") {
		trimmedExpresion := strings.TrimSuffix(expression, "*")
		for _, i := range indiceResult {
			if strings.HasPrefix(i.Name, trimmedExpresion) {
				indexResult = append(indexResult, i.Name)
			}
		}
	} else {
		for _, i := range indiceResult {
			if strings.Compare(i.Name, expression) == 0 {
				indexResult = append(indexResult, i.Name)
			}
		}
	}

	return indexResult
}

func (r *NameExpressionResolver) ConcreteSingleIndex(clusterState state.ClusterState, expression string) state.Index {
	if indices := r.ConcreteIndices(clusterState, expression); len(indices) == 0 {
		return state.Index{}
	} else {
		return indices[0]
	}
}

func (r *NameExpressionResolver) ConcreteIndices(clusterState state.ClusterState, expression string) []state.Index {
	var indicesResult []state.Index

	if strings.HasSuffix(expression, "*") {
		trimmedExpression := strings.TrimSuffix(expression, "*")
		for k, v := range clusterState.Metadata.Indices {
			if strings.HasPrefix(k, trimmedExpression) {
				indicesResult = append(indicesResult, v.Index)
			}
		}
	} else {
		for k, v := range clusterState.Metadata.Indices {
			if strings.HasPrefix(k, expression) {
				indicesResult = append(indicesResult, v.Index)
			}
		}
	}

	if indexAbstraction, existing := clusterState.Metadata.IndicesLookup[expression]; existing {
		indicesResult = append(indicesResult, indexAbstraction.WriteIndex.Index)
	}

	return indicesResult
}
