package index

import (
	"github.com/actumn/searchgoose/errors"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/index/scorch"
	"github.com/blevesearch/bleve/mapping"
	"log"
	"time"
)

type Shard struct {
	engine bleve.Index
}

func NewShard(shardPath string, mapping mapping.IndexMapping) *Shard {
	index, err := bleve.NewUsing(shardPath, mapping, scorch.Name, scorch.Name, map[string]interface{}{
		"create_if_missing": true,
		"error_if_exists":   false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	return &Shard{
		engine: index,
	}
}

func (s *Shard) Index(id string, fields map[string]interface{}) error {
	if err := s.engine.Index(id, fields); err != nil {
		return err
	}
	return nil
}

func (s *Shard) Delete(id string) error {
	if err := s.engine.Delete(id); err != nil {
		return err
	}
	return nil
}

func (s *Shard) Get(id string) (map[string]interface{}, error) {
	doc, err := s.engine.Document(id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, errors.ErrNotFound
	}

	fields := make(map[string]interface{}, 0)
	for _, f := range doc.Fields {
		var v interface{}
		switch field := f.(type) {
		case *document.TextField:
			v = string(field.Value())
		case *document.NumericField:
			n, err := field.Number()
			if err == nil {
				v = n
			}
		case *document.DateTimeField:
			d, err := field.DateTime()
			if err == nil {
				v = d.Format(time.RFC3339Nano)
			}
		}
		existing, existed := fields[f.Name()]
		if existed {
			switch existing := existing.(type) {
			case []interface{}:
				fields[f.Name()] = append(existing, v)
			case interface{}:
				arr := make([]interface{}, 2)
				arr[0] = existing
				arr[1] = v
				fields[f.Name()] = arr
			}
		} else {
			fields[f.Name()] = v
		}
	}

	return fields, nil
}

func (s *Shard) Search(searchRequest *bleve.SearchRequest) (*bleve.SearchResult, error) {
	searchResult, err := s.engine.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	return searchResult, nil
}