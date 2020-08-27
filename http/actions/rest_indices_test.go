package actions

import (
	"fmt"
	"testing"
)

func TestRestPutIndices_Handle(t *testing.T) {
	restPut := RestPutIndices{}
	req := &RestRequest{
		PathParams: map[string]string{
			"index": "test",
		},
		Body: []byte(`
		{
			"settings": {
				"number_of_shards": 3,
				"number_of_replicas": 2
			},
			"mappings": {
				"properties": {
					"field1": {
						"type": "text"
					}
				}
			}
		}`),
	}

	restPut.Handle(req,
		func(response RestResponse) {
			fmt.Println(response.StatusCode)
			fmt.Println(response)
		},
	)
}
