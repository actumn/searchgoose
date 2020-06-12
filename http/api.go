package http

import (
	"github.com/actumn/searchgoose/index"
	"github.com/blevesearch/bleve/mapping"
	"github.com/gin-gonic/gin"
)

type Bootstrap struct {
	i *index.Index
	r *gin.Engine
}

func New() *Bootstrap {
	r := gin.Default()
	indexMapping := mapping.NewIndexMapping()
	i, _ := index.NewIndex("./examples", indexMapping)

	r.GET("/", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"name": "searchgoose",
			"version": map[string]interface{}{
				"number": "0.0.0",
			},
		})
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
