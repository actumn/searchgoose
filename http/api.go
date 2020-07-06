package http

import (
	"github.com/actumn/searchgoose/cluster"
	"github.com/actumn/searchgoose/index"
	"github.com/blevesearch/bleve/mapping"
	"github.com/gin-gonic/gin"
)

type Bootstrap struct {
	i *index.Index
	r *gin.Engine
}

func misc(r *gin.Engine) {
	r.GET("/", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"name": "searchgoose",
			"version": map[string]interface{}{
				"number": "0.0.0",
			},
		})
	})
	r.GET("/_nodes", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"nodes": map[string]interface{}{
				cluster.GenerateNodeId(): map[string]interface{}{
					"ip":      "127.0.0.1",
					"version": "7.8.0",
					"http": map[string]interface{}{
						"public_address": "127.0.0.1:8080",
					},
				},
			},
		})
	})
	r.GET("/_xpack", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"license": map[string]interface{}{
				"uid":    "d0309419-2f93-4b56-95e1-97c0c6415956",
				"type":   "basic",
				"mode":   "basic",
				"status": "active",
			},
		})
	})

	r.NoRoute(func(context *gin.Context) {
		context.JSON(404, gin.H{
			"error": map[string]interface{}{
				"root_capuse": []map[string]interface{}{
					{
						"type":          "index_not_found_exception",
						"reason":        "no such index [.kibana]",
						"resource.type": "index_or_alias",
						"resource.id":   ".kibana",
						"index_uuid":    "_na_",
						"index":         ".kibana",
					},
				},
				"type":          "index_not_found_exception",
				"reason":        "no such index [.kibana]",
				"resource.type": "index_or_alias",
				"resource.id":   ".kibana",
				"index_uuid":    "_na_",
				"index":         ".kibana",
			},
			"status": 404,
		})
	})
}

func New() *Bootstrap {
	r := gin.Default()
	indexMapping := mapping.NewIndexMapping()
	i, _ := index.NewIndex("./examples", indexMapping)

	misc(r)

	r.GET("/_cat/templates/:template", func(context *gin.Context) {
		context.JSON(200, []interface{}{})
	})

	r.GET("/_doc/:id", func(context *gin.Context) {
		id := context.Param("id")
		doc, err := i.Get(id)
		if err != nil {
			context.JSON(404, gin.H{
				"message": err.Error(),
			})
			return
		}

		context.JSON(200, doc)
	})

	r.PUT("/_doc/:id", func(context *gin.Context) {
		id := context.Param("id")
		var doc map[string]interface{}
		if err := context.Bind(&doc); err != nil {
			context.JSON(500, gin.H{
				"code":    500,
				"message": err.Error(),
			})
			return
		}
		if err := i.Index(id, doc); err != nil {
			context.JSON(500, gin.H{
				"code":    500,
				"message": err.Error(),
			})
			return
		}
	})

	r.DELETE("/_doc/:id", func(context *gin.Context) {
		id := context.Param("id")
		err := i.Delete(id)
		if err != nil {
			context.JSON(500, gin.H{
				"message": err.Error(),
			})
		}
		context.JSON(200, gin.H{
			"message": "OK",
		})
	})

	return &Bootstrap{
		r: r,
	}
}

func (b *Bootstrap) Start() error {
	return b.r.Run()
}
