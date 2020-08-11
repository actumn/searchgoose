package actions

type RestXpack struct{}

func (h *RestXpack) Handle(r *RestRequest) (interface{}, error) {
	return map[string]interface{}{
		"license": map[string]interface{}{
			"uid":    "d0309419-2f93-4b56-95e1-97c0c6415956",
			"type":   "basic",
			"mode":   "basic",
			"status": "active",
		},
	}, nil
}
