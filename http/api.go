package http

import (
	"bytes"
	"encoding/json"
	"github.com/actumn/searchgoose/http/actions"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/valyala/fasthttp"
	"log"
)

var (
	strContentType     = []byte("Content-Type")
	strApplicationJSON = []byte("application/json")
)

func requestFromCtx(ctx *fasthttp.RequestCtx) actions.RestRequest {
	request := actions.RestRequest{
		Path:        string(ctx.Path()),
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
	log.Println(request.Path)
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
				log.Println(err)
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
				log.Println(err)
			}
		})
		return
	}
}

type Bootstrap struct {
	s *fasthttp.Server
}

func New(clusterService *cluster.Service) *Bootstrap {
	//indexMapping := mapping.NewIndexMapping()
	//i, _ := index.NewIndex("./examples", indexMapping)
	//
	//r.GET("/_doc/:id", func(context *gin.Context) {
	//	id := context.Param("id")
	//	doc, err := i.Get(id)
	//	if err != nil {
	//		context.JSON(404, gin.H{
	//			"message": err.Error(),
	//		})
	//		return
	//	}
	//
	//	context.JSON(200, doc)
	//})
	//
	//r.PUT("/_doc/:id", func(context *gin.Context) {
	//	id := context.Param("id")
	//	var doc map[string]interface{}
	//	if err := context.Bind(&doc); err != nil {
	//		context.JSON(500, gin.H{
	//			"code":    500,
	//			"message": err.Error(),
	//		})
	//		return
	//	}
	//	if err := i.Index(id, doc); err != nil {
	//		context.JSON(500, gin.H{
	//			"code":    500,
	//			"message": err.Error(),
	//		})
	//		return
	//	}
	//})
	//
	//r.DELETE("/_doc/:id", func(context *gin.Context) {
	//	id := context.Param("id")
	//	err := i.Delete(id)
	//	if err != nil {
	//		context.JSON(500, gin.H{
	//			"message": err.Error(),
	//		})
	//	}
	//	context.JSON(200, gin.H{
	//		"message": "OK",
	//	})
	//})

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
		actions.PUT: &actions.RestPutIndex{},
	})
	c.pathTrie.insert("/{index}", actions.MethodHandlers{
		actions.GET: &actions.RestGetIndex{},
	})
	c.pathTrie.insert("/{index}", actions.MethodHandlers{
		actions.DELETE: &actions.RestDeleteIndex{},
	})
	c.pathTrie.insert("/{index}", actions.MethodHandlers{
		actions.HEAD: &actions.RestHeadIndex{},
	})
	c.pathTrie.insert("/{index}/_doc", actions.MethodHandlers{
		actions.POST: &actions.RestIndexDoc{},
	})
	c.pathTrie.insert("/{index}/_doc/{id}", actions.MethodHandlers{
		actions.PUT: &actions.RestIndexDocId{},
	})
	c.pathTrie.insert("/{index}/_doc/{id}", actions.MethodHandlers{
		actions.GET: &actions.RestGetDoc{},
	})
	c.pathTrie.insert("/{index}/_doc/{id}", actions.MethodHandlers{
		actions.HEAD: &actions.RestHeadDoc{},
	})
	c.pathTrie.insert("/{index}/_doc/{id}", actions.MethodHandlers{
		actions.DELETE: &actions.RestDeleteDoc{},
	})
	c.pathTrie.insert("/{index}/_search", actions.MethodHandlers{
		actions.GET: &actions.RestIndexSearch{},
	})

	s := &fasthttp.Server{
		Handler: c.HandleFastHTTP,
	}

	return &Bootstrap{
		s: s,
	}
}

func (b *Bootstrap) Start() error {
	return b.s.ListenAndServe(":8080")
}
