package http

import (
	"github.com/gin-gonic/gin"
)

func Api(r *gin.Engine) {
	v1 := r.Group("v1")

	v1.GET("/ping", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"message": "pong",
		})
	})

}
