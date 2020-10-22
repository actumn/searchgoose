package actions

import (
	"fmt"
	"testing"
)

func TestRestPostIndexAlias_Handle(t *testing.T) {
	// Arrange
	postIndexAlias := RestPostIndexAlias{}
	req := &RestRequest{
		Body: []byte(`{
   "actions":[
      {
         "remove":{
            "index":".kibana_task_manager",
            "alias":".kibana_task_manager"
         }
      },
      {
         "add":{
            "index":".kibana_task_manager_1",
            "alias":".kibana_task_manager"
         }
      }
  ]
}`),
	}

	// Action
	postIndexAlias.Handle(req,
		func(response RestResponse) {
			fmt.Println(response.StatusCode)
			fmt.Println(response)
		},
	)
}
