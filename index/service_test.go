package index

import (
	"github.com/actumn/searchgoose/state"
	"testing"
)

func TestService_UpdateMapping(t *testing.T) {
	// Arrange
	indexService := NewService()
	indexMetadata := state.IndexMetadata{
		Mapping: map[string]state.MappingMetadata{
			"_doc": {
				Type: "_doc",
				Source: []byte(`{
				"properties": {
					"field1": {
						"type": "text"
					}
				}
			}`),
			},
		},
	}

	// Action
	indexService.UpdateMapping(indexMetadata)

	// Assert
	//fmt.Println(indexService.indexMapping)
}
