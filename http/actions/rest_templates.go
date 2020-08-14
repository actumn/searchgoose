package actions

type RestTemplates struct{}

func (h *RestTemplates) Handle(r *RestRequest, reply ResponseListener) {
	reply(RestResponse{
		StatusCode: 200,
		Body:       []interface{}{},
	})
}
