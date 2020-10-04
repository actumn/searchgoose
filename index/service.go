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

	// TODO :: implements more mapping types (numeric, geo, datetime, boolean, sub-document, ...)
	properties := indexMapping["properties"].(map[string]interface{})
	for field, fieldProps := range properties {
		props := fieldProps.(map[string]interface{})

		switch props["type"] {
		case "text":
			docMapping.AddFieldMappingsAt(field, mapping.NewTextFieldMapping())
		case "integer":
			docMapping.AddFieldMappingsAt(field, mapping.NewNumericFieldMapping())
		case "float":
			docMapping.AddFieldMappingsAt(field, mapping.NewNumericFieldMapping())
		case "date":
			docMapping.AddFieldMappingsAt(field, mapping.NewDateTimeFieldMapping())
		case "boolean":
			docMapping.AddFieldMappingsAt(field, mapping.NewBooleanFieldMapping())
		case "geo_point":
			docMapping.AddFieldMappingsAt(field, mapping.NewGeoPointFieldMapping())
		}
	}
}

func (s *Service) CreateShard(shardId int) {
	shard := NewShard("./data/"+s.uuid+"/"+strconv.Itoa(shardId), s.indexMapping)
	s.shards[shardId] = shard
}

func (s *Service) Shard(shardId int) (*Shard, bool) {
	shard, ok := s.shards[shardId]
	return shard, ok
}
