package main

import (
	"github.com/actumn/searchgoose/http"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	http.Api(r)
	err := r.Run()
	if err != nil {
		panic(err)
	}
}
