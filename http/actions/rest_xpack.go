package actions

type RestXpack struct{}

func (h *RestXpack) Handle(r *RestRequest, reply ResponseListener) {
	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"license": map[string]interface{}{
				"uid":    "d0309419-2f93-4b56-95e1-97c0c6415956",
				"type":   "basic",
				"mode":   "basic",
				"status": "active",
			},
		},
	})
}
