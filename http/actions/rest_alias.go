package actions

type RestGetIndexAlias struct {
}

func NewRestGetIndexAlias() *RestGetIndexAlias {
	return &RestGetIndexAlias{}
}

func (h *RestGetIndexAlias) Handle(r *RestRequest, reply ResponseListener) {
	name := r.PathParams["name"]

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			name: map[string]interface{}{},
		},
	})
}

type RestPostIndexAlias struct {
}

/*
{
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
}
*/
func NewRestPostIndexAlias() *RestPostIndexAlias {
	return &RestPostIndexAlias{}
}

func (h *RestPostIndexAlias) Handle(r *RestRequest, reply ResponseListener) {

}
