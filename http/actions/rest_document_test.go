package actions

import (
	"fmt"
	"testing"
)

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
