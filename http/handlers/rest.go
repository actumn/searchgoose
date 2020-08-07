package handlers

type RestMethod int

const (
	GET RestMethod = iota
	POST
	PUT
	DELETE
)

type MethodHandlers map[RestMethod]RestHandler

type RestRequest struct {
	Path        string
	Method      RestMethod
	PathParams  map[string]string
	QueryParams map[string][]byte
	Body        []byte
}

type RestHandler interface {
	Handle(r *RestRequest) (interface{}, error)
}
