package index

import (
	"github.com/blevesearch/bleve/mapping"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestIndex_Get(t *testing.T) {
	indexMapping := mapping.NewIndexMapping()
	s := NewShard("/examples", indexMapping)
	id := "test"
	if doc, err := s.Get(id); err != nil {
		logrus.Fatal(err)
	} else {
		logrus.Info(doc)
	}
}

func TestIndex_Index(t *testing.T) {
	indexMapping := mapping.NewIndexMapping()
	s := NewShard("/examples", indexMapping)
	id := "test"
	doc := map[string]interface{}{}
	if err := s.Index(id, doc); err != nil {
		logrus.Fatal(err)
	}
}

func TestIndex_Delete(t *testing.T) {
	indexMapping := mapping.NewIndexMapping()
	s := NewShard("/examples", indexMapping)
	id := "test"
	if err := s.Delete(id); err != nil {
		logrus.Fatal(err)
	}
}
