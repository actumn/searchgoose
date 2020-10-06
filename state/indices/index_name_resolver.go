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

func (r *NameExpressionResolver) ConcreteIndices(clusterState state.ClusterState, expression string) []state.Index {
	var indiceResult []state.Index

	if strings.HasSuffix(expression, "*") {
		trimmedExpression := strings.TrimSuffix(expression, "*")
		for k, v := range clusterState.Metadata.Indices {
			if strings.HasPrefix(k, trimmedExpression) {
				indiceResult = append(indiceResult, v.Index)
			}
		}
	} else {
		for k, v := range clusterState.Metadata.Indices {
			if strings.HasPrefix(k, expression) {
				indiceResult = append(indiceResult, v.Index)
			}
		}
	}

	return indiceResult
}
