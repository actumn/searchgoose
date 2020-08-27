package index

import (
	"github.com/blevesearch/bleve/mapping"
	"log"
	"testing"
)

func TestIndex_Get(t *testing.T) {
	indexMapping := mapping.NewIndexMapping()
	s := NewShard("/examples", indexMapping)
	id := "test"
	if doc, err := s.Get(id); err != nil {
		log.Fatalln(err)
	} else {
		log.Println(doc)
	}
}

func TestIndex_Index(t *testing.T) {
	indexMapping := mapping.NewIndexMapping()
	s := NewShard("/examples", indexMapping)
	id := "test"
	doc := map[string]interface{}{}
	if err := s.Index(id, doc); err != nil {
		log.Fatalln(err)
	}
}

func TestIndex_Delete(t *testing.T) {
	indexMapping := mapping.NewIndexMapping()
	s := NewShard("/examples", indexMapping)
	id := "test"
	if err := s.Delete(id); err != nil {
		log.Fatalln(err)
	}
}
