package actions

import (
	"fmt"
	"github.com/actumn/searchgoose/state"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIndexRequest_ToBytes(t *testing.T) {
	// Arrange
	req := indexRequest{
		Index:  "index",
		Id:     "test-id",
		Source: []byte("asdfmasklfmklsmfkl"),
		ShardId: state.ShardId{
			Index: state.Index{
				Name: "index",
				Uuid: "jujsyy234lds,",
			},
			ShardId: 1,
		},
	}

	// Action
	bytes := req.toBytes()
	parsed := indexRequestFromBytes(bytes)

	// Assert
	assert.Equal(t, req.Source, parsed.Source)
}

func TestRestGetDoc_Handle(t *testing.T) {
	// Arrange
	restGet := RestGetDoc{}
	req := &RestRequest{
		Path: "/my-index-000001/_doc/0",
		PathParams: map[string]string{
			"index": "my-index-000001",
			"id":    "0",
		},
	}

	// Action
	restGet.Handle(req,
		func(response RestResponse) {
			fmt.Println(response.StatusCode)
			fmt.Println(response)
		},
	)
}

func TestRestIndexDocId_Handle(t *testing.T) {
	// Arrange
	restPut := RestIndexDocId{}
	req := &RestRequest{
		Path: "/my-index-000001/_doc/0",
		PathParams: map[string]string{
			"index": "my-index-000001",
			"id":    "0",
		},

		Body: []byte(`
		{
			"@timestamp": "2099-11-15T13:12:00",
			"message": "GET /search HTTP/1.1 200 1070000",
			"user": {
				"id": "kimchy"
			}
		}`),
	}

	// Action
	restPut.Handle(req,
		func(response RestResponse) {
			fmt.Println(response.StatusCode)
			fmt.Println(response)
		},
	)
}
