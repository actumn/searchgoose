package handlers

type RestTemplates struct{}

func (h *RestTemplates) Handle(r *RestRequest) (interface{}, error) {
	return []interface{}{}, nil
}
