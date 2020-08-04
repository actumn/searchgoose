package handlers

type RestRequest struct {
}

type RestHandler interface {
	Handle(r *RestRequest) (interface{}, error)
}
