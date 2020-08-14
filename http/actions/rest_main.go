package actions

type RestMain struct{}

func (h *RestMain) Handle(r *RestRequest, reply ResponseListener) {
	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"name": "searchgoose",
			"version": map[string]interface{}{
				"number": "0.0.0",
			},
		},
	})
}
