package main

import (
	"github.com/actumn/searchgoose/http"
	"log"
)

func main() {
	b := http.New()
	log.Println("start server...")
	if err := b.Start(); err != nil {
		panic(err)
	}
}
