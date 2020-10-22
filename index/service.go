package index

import (
	"encoding/json"
	"github.com/actumn/searchgoose/state"
	"github.com/blevesearch/bleve/mapping"
	"github.com/sirupsen/logrus"
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
		logrus.Fatal(err)
	}

	// TODO :: implements more mapping types (numeric, geo, datetime, boolean, sub-document, ...)
	properties := indexMapping["properties"].(map[string]interface{})
	for field, fieldProps := range properties {
		props := fieldProps.(map[string]interface{})

		switch props["type"] {
		case "text", "keyword":
			docMapping.AddFieldMappingsAt(field, mapping.NewTextFieldMapping())
		case "long", "integer", "short", "byte", "double", "float", "half_float", "scaled_float":
			docMapping.AddFieldMappingsAt(field, mapping.NewNumericFieldMapping())
		case "date":
			docMapping.AddFieldMappingsAt(field, mapping.NewDateTimeFieldMapping())
		case "boolean":
			docMapping.AddFieldMappingsAt(field, mapping.NewBooleanFieldMapping())
		case "geo_point":
			docMapping.AddFieldMappingsAt(field, mapping.NewGeoPointFieldMapping())
		//case "binary":
		//	docMapping.AddFieldMappingsAt(field, NewBinaryFieldMapping())
		//case "range":
		//	docMapping.AddFieldMappingsAt(field, NewRangeFieldMapping())
		case "object":
			docMapping.AddFieldMappingsAt(field, NewObjectFieldMapping())
		case "nested":
			docMapping.AddFieldMappingsAt(field, NewNestedFieldMapping())
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
