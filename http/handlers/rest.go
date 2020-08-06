package handlers

type RestRequest struct {
	Path        string
	Method      []byte
	PathParams  map[string]string
	QueryParams map[string][]byte
	Body        []byte
}

type RestHandler interface {
	Handle(r *RestRequest) (interface{}, error)
}
