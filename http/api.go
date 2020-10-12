package http

import (
	"bytes"
	"encoding/json"
	"github.com/actumn/searchgoose/http/actions"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

var (
	strContentType     = []byte("Content-Type")
	strApplicationJSON = []byte("application/json")
)

func requestFromCtx(ctx *fasthttp.RequestCtx) actions.RestRequest {
	request := actions.RestRequest{
		Path:        string(ctx.Path()),
		Header:      map[string][]byte{},
		QueryParams: map[string][]byte{},
		Body:        ctx.Request.Body(),
	}
	if bytes.Compare(ctx.Method(), []byte("GET")) == 0 {
		request.Method = actions.GET
	} else if bytes.Compare(ctx.Method(), []byte("POST")) == 0 {
		request.Method = actions.POST
	} else if bytes.Compare(ctx.Method(), []byte("PUT")) == 0 {
		request.Method = actions.PUT
	} else if bytes.Compare(ctx.Method(), []byte("DELETE")) == 0 {
		request.Method = actions.DELETE
	} else if bytes.Compare(ctx.Method(), []byte("HEAD")) == 0 {
		request.Method = actions.HEAD
	}
	ctx.Request.Header.VisitAll(func(key, value []byte) {
		request.Header[string(key)] = value
	})
	ctx.QueryArgs().VisitAll(func(key, value []byte) {
		request.QueryParams[string(key)] = value
	})

	return request
}

type RequestController struct {
	pathTrie *pathTrie
}

func (c *RequestController) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	request := requestFromCtx(ctx)
	if request.Path == "/favicon.ico" {
		return
	}
	logrus.Info(string(ctx.Method()), " ", request.Path, " ", string(ctx.Request.Body()))
	allHandlers := c.pathTrie.retrieveAll(request.Path)
	for {
		h, params, err := allHandlers()
		request.PathParams = params
		if err != nil {
			ctx.Response.Header.SetCanonical(strContentType, strApplicationJSON)
			ctx.Response.SetStatusCode(400)
			if err := json.NewEncoder(ctx).Encode(map[string]string{
				"msg": "no route",
			}); err != nil {
				logrus.Error(err)
			}
			return
		}
		methodHandlers, ok := h.(actions.MethodHandlers)
		if !ok {
			continue
		}

		handler, ok := methodHandlers[request.Method]
		if !ok {
			continue
		}

		handler.Handle(&request, func(response actions.RestResponse) {
			ctx.Response.SetStatusCode(response.StatusCode)
			ctx.Response.Header.SetCanonical(strContentType, strApplicationJSON)
			if err := json.NewEncoder(ctx).Encode(response.Body); err != nil {
				logrus.Info(err)
			}
		})
		return
	}
}

type Bootstrap struct {
	s *fasthttp.Server
}

func New(
	clusterService *cluster.Service,
	clusterMetadataCreateIndexService *cluster.MetadataCreateIndexService,
	clusterMetadataDeleteIndexService *cluster.MetadataDeleteIndexService,
	indicesService *indices.Service,
	transportService *transport.Service,
	indexNameExpressionResolver *indices.NameExpressionResolver,
) *Bootstrap {
	c := RequestController{}
	c.pathTrie = newPathTrie()
	c.pathTrie.insert("/", actions.MethodHandlers{
		actions.GET: &actions.RestMain{},
	})
	c.pathTrie.insert("/_cat/templates", actions.MethodHandlers{
		actions.GET: &actions.RestTemplates{},
	})
	c.pathTrie.insert("/_cat/templates/{name}", actions.MethodHandlers{
		actions.GET: &actions.RestTemplates{},
	})
	c.pathTrie.insert("/_nodes", actions.MethodHandlers{
		actions.GET: &actions.RestNodes{
			ClusterService: clusterService,
		},
	})
	c.pathTrie.insert("/_xpack", actions.MethodHandlers{
		actions.GET: &actions.RestXpack{},
	})
	c.pathTrie.insert("/{index}", actions.MethodHandlers{
		actions.GET:    actions.NewRestGetIndex(clusterService, indexNameExpressionResolver),
		actions.PUT:    actions.NewRestPutIndex(clusterMetadataCreateIndexService),
		actions.DELETE: actions.NewRestDeleteIndex(clusterService, indexNameExpressionResolver, clusterMetadataDeleteIndexService),
		actions.HEAD:   &actions.RestHeadIndex{},
	})
	c.pathTrie.insert("/{index}/_doc", actions.MethodHandlers{
		actions.POST: actions.NewRestIndexDoc(clusterService, indicesService, transportService),
	})
	c.pathTrie.insert("/{index}/_doc/{id}", actions.MethodHandlers{
		actions.GET:    actions.NewRestGetDoc(clusterService, indicesService, transportService),
		actions.PUT:    actions.NewRestIndexDocId(clusterService, indicesService, transportService),
		actions.DELETE: actions.NewRestDeleteDoc(clusterService, indicesService, transportService),
		actions.HEAD:   &actions.RestHeadDoc{},
	})
	c.pathTrie.insert("/{index}/_search", actions.MethodHandlers{
		actions.GET: actions.NewRestSearch(clusterService, indicesService, transportService),
	})
	//c.pathTrie.insert("/{index}/_bulk", actions.MethodHandlers{
	//	actions.POST: ,
	//	actions.PUT: ,
	//})

	s := &fasthttp.Server{
		Handler: c.HandleFastHTTP,
	}

	return &Bootstrap{
		s: s,
	}
}

func (b *Bootstrap) Start(port string) error {
	return b.s.ListenAndServe(port)
}
