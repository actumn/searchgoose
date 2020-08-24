package index

import (
	"github.com/blevesearch/bleve/mapping"
	"log"
	"testing"
)

func TestIndex_Get(t *testing.T) {
	indexMapping := mapping.NewIndexMapping()
	i, _ := NewIndex("./examples", indexMapping)
	id := "test"
	if doc, err := i.Get(id); err != nil {
		log.Fatalln(err)
	} else {
		log.Println(doc)
	}
}

func TestIndex_Index(t *testing.T) {
	indexMapping := mapping.NewIndexMapping()
	i, _ := NewIndex("./examples", indexMapping)
	id := "test"
	doc := map[string]interface{}{}
	if err := i.Index(id, doc); err != nil {
		log.Fatalln(err)
	}
}

func TestIndex_Delete(t *testing.T) {
	indexMapping := mapping.NewIndexMapping()
	i, _ := NewIndex("./examples", indexMapping)
	id := "test"
	if err := i.Delete(id); err != nil {
		log.Fatalln(err)
	}
}
