package http

import (
	"encoding/json"
	"github.com/actumn/searchgoose/http/handlers"
	"github.com/valyala/fasthttp"
	"log"
)

var (
	strContentType     = []byte("Content-Type")
	strApplicationJSON = []byte("application/json")
)

type RequestController struct {
	pathTrie *pathTrie
}

func (c *RequestController) init() {
	c.pathTrie = newPathTrie()
	c.pathTrie.insert("/", &handlers.RestMain{})
	c.pathTrie.insert("/_nodes", &handlers.RestNodes{})
	c.pathTrie.insert("/_xpack", &handlers.RestXpack{})
}

func (c *RequestController) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	if path == "/favicon.ico" {
		return
	}
	log.Println(path)
	allHandlers := c.pathTrie.retrieveAll(path)
	for {
		h, _, err := allHandlers()
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
		handler, ok := h.(handlers.RestHandler)
		if ok {
			request := handlers.RestRequest{}
			response, err := handler.Handle(&request)
			if err == nil {
				ctx.Response.Header.SetCanonical(strContentType, strApplicationJSON)
				ctx.Response.SetStatusCode(200)
				if err := json.NewEncoder(ctx).Encode(response); err != nil {
					log.Println(err)
				}
			} else {
				ctx.Response.Header.SetCanonical(strContentType, strApplicationJSON)
				ctx.Response.SetStatusCode(500)
				if err := json.NewEncoder(ctx).Encode(map[string]string{
					"msg": err.Error(),
				}); err != nil {
					log.Println(err)
				}
			}
			break
		}
	}
}

type Bootstrap struct {
	s *fasthttp.Server
}

func New() *Bootstrap {
	//r := gin.Default()
	//indexMapping := mapping.NewIndexMapping()
	//i, _ := index.NewIndex("./examples", indexMapping)
	//
	//r.GET("/_cat/templates/:template", func(context *gin.Context) {
	//	context.JSON(200, []interface{}{})
	//})
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
	c.init()
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
