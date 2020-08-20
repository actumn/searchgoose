package actions

type RestMethod int

const (
	GET RestMethod = iota
	POST
	PUT
	DELETE
	HEAD
)

type MethodHandlers map[RestMethod]RestHandler

type RestRequest struct {
	Path        string
	Method      RestMethod
	PathParams  map[string]string
	QueryParams map[string][]byte
	Body        []byte
}

type RestResponse struct {
	StatusCode int
	Body       interface{}
}

type ResponseListener func(response RestResponse)

type RestHandler interface {
	Handle(r *RestRequest, reply ResponseListener)
}
