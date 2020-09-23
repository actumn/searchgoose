package actions

import (
	"fmt"
	"testing"
)

func TestRestIndexSearch_Handle(t *testing.T) {
	restSearch := RestSearch{}
	req := &RestRequest{
		Method: GET,
		Path:   "/my-index-000001/_search",
		PathParams: map[string]string{
			"index": "my-index-000001",
		},
	}

	restSearch.Handle(req,
		func(response RestResponse) {
			fmt.Println(response.StatusCode)
			fmt.Println(response)
		},
	)
}
func TestRestIndexSearch_Handle2(t *testing.T) {
	restSearch := RestSearch{}
	req := &RestRequest{
		Method: POST,
		Path:   "/my-index-000001/_search",
		PathParams: map[string]string{
			"index": "my-index-000001",
		},
		Body: []byte(`
		{
			"query": {
				"match_all": {}
			}
		}`),
	}

	restSearch.Handle(req,
		func(response RestResponse) {
			fmt.Println(response.StatusCode)
			fmt.Println(response)
		},
	)
}

func TestRestIndexSearch_Handle3(t *testing.T) {
	restSearch := RestSearch{}
	req := &RestRequest{
		Method: POST,
		Path:   "/my-index-000001/_search",
		PathParams: map[string]string{
			"index": "my-index-000001",
		},
		Body: []byte(`
		{
			"query": {
				"term": {
					"user.id": "kimchy"
				}
			}
		}`),
	}

	restSearch.Handle(req,
		func(response RestResponse) {
			fmt.Println(response.StatusCode)
			fmt.Println(response)
		},
	)
}
