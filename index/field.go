package index

import "github.com/blevesearch/bleve/mapping"

func NewObjectFieldMapping() *mapping.FieldMapping {
	return &mapping.FieldMapping{
		Type:         "text",
		Store:        true,
		Index:        true,
		IncludeInAll: true,
		DocValues:    true,
	}
}

func NewNestedFieldMapping() *mapping.FieldMapping {
	return &mapping.FieldMapping{
		Type:         "text",
		Store:        true,
		Index:        true,
		IncludeInAll: true,
		DocValues:    true,
	}
}

/*func NewRangeFieldMapping() *mapping.FieldMapping {
	return &mapping.FieldMapping{
		Type:         "range",
		Store:        true,
		Index:        true,
		IncludeInAll: true,
		DocValues:    true,
	}
}*/

//func NewBinaryFieldMapping() *mapping.FieldMapping {
//	return &mapping.FieldMapping{
//		Type:         "binary",
//		Store:        true,
//		Index:        true,
//		IncludeInAll: true,
//		DocValues:    true,
//	}
//}
