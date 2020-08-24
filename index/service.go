package index

import "github.com/blevesearch/bleve/mapping"

type Service struct {
	shards       map[int]*Shard
	indexMapping mapping.IndexMapping
}

func (s *Service) UpdateMapping() {

}

func (s *Service) CreateShard() {
	//shard := NewShard()
}
