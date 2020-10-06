package index

import (
	"github.com/actumn/searchgoose/state"
	"testing"
)

func TestService_UpdateMapping(t *testing.T) {
	// Arrange
	indexService := NewService("test")
	indexMetadata := state.IndexMetadata{
		Mapping: map[string]state.MappingMetadata{
			"_doc": {
				Type: "_doc",
				Source: []byte(`{
				"properties": {
					"text_field": {
						"type": "text"
					},
					"integer_field": {
						"type": "integer"
					},
					"float_field": {
						"type": "float"
					},
					"date_field": {
						"type": "date"
					},
					"boolean_field": {
						"type": "boolean"
					},
					"geo_point_field": {
						"type": "geo_point"
					},
					"object_field": {
						"properties": {
							"age": {
								"type": "integer"
							},
							"name": {
								"properties": {
									"first": {
										"type": "text"
									},
									"last": {
										"type": "text"
									}
								}
							}
						}
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
