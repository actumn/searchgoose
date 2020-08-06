package handlers

type RestMain struct{}

func (h *RestMain) Handle(r *RestRequest) (interface{}, error) {
	return map[string]interface{}{
		"name": "searchgoose",
		"version": map[string]interface{}{
			"number": "0.0.0",
		},
	}, nil
}
