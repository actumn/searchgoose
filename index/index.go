package index

import (
	"github.com/actumn/searchgoose/errors"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/index/scorch"
	"github.com/blevesearch/bleve/mapping"
	"os"
	"time"
)

type Index struct {
	index        bleve.Index
	indexMapping *mapping.IndexMapping
}

func NewIndex(dir string, indexMapping *mapping.IndexMapping) (*Index, error) {
	var index bleve.Index

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		index, err = bleve.NewUsing(dir, *indexMapping, scorch.Name, scorch.Name, nil)
		if err != nil {
			return nil, err
		}
	} else {
		index, err = bleve.OpenUsing(dir, map[string]interface{}{
			"create_if_missing": false,
			"error_if_exists":   false,
		})
		if err != nil {
			return nil, err
		}
	}

	return &Index{
		index:        index,
		indexMapping: indexMapping,
	}, nil
}

func (i *Index) Close() error {
	if err := i.index.Close(); err != nil {
		return err
	}
	return nil
}

func (i *Index) Get(id string) (map[string]interface{}, error) {
	doc, err := i.index.Document(id)
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
