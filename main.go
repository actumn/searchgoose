package main

import (
	"github.com/actumn/searchgoose/http"
)

func main() {
	b := http.New()
	if err := b.Start(); err != nil {
		panic(err)
	}
}
