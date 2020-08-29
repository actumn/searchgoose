package index

import (
	"encoding/json"
	"github.com/actumn/searchgoose/state"
	"github.com/blevesearch/bleve/mapping"
	"log"
	"strconv"
)

type Service struct {
	uuid         string
	shards       map[int]*Shard
	indexMapping *mapping.IndexMappingImpl
}

func NewService(uuid string) *Service {
	return &Service{
		uuid:         uuid,
		shards:       map[int]*Shard{},
		indexMapping: mapping.NewIndexMapping(),
	}
}

func (s *Service) UpdateMapping(metadata state.IndexMetadata) {
	mappingMetadata := metadata.Mapping["_doc"]
	docMapping := mapping.NewDocumentMapping()
	s.indexMapping.AddDocumentMapping("_doc", docMapping)

	var indexMapping map[string]interface{}
	if err := json.Unmarshal(mappingMetadata.Source, &indexMapping); err != nil {
		log.Fatalln(err)
	}

	properties := indexMapping["properties"].(map[string]interface{})
	for field, fieldProps := range properties {
		props := fieldProps.(map[string]interface{})
		if props["type"] == "text" {
			docMapping.AddFieldMappingsAt(field, mapping.NewTextFieldMapping())
		}
	}
}

func (s *Service) CreateShard() {
	shardId := 0
	shard := NewShard("./data/"+s.uuid+"/"+strconv.Itoa(shardId), s.indexMapping)
	s.shards[shardId] = shard
}
