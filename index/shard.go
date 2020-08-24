package index

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

type Shard struct {
	storage bleve.Index
}

func NewShard(mapping mapping.IndexMapping) *Shard {
	return &Shard{}
}

func (s *Shard) Index() {

}

func (s *Shard) Get() {

}

func (s *Shard) Delete() {

}
