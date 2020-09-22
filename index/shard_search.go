package index

import (
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"strings"
)

func (s *Shard) SearchTypeMatch(searchType interface{}) (query *query.MatchQuery) {
	m := searchType.(map[string]interface{})
	var field, message string
	for key, value := range m {
		field = key
		message = value.(string)
	}
	queryString := field + ":\"" + message + "\""

	query = bleve.NewMatchQuery(queryString)
	return query
}

func (s *Shard) SearchTypeMatchPhrase(searchType interface{}) (query *query.MatchPhraseQuery) {
	m := searchType.(map[string]interface{})
	var field, message string
	for key, value := range m {
		field = key
		message = value.(string)
	}
	queryString := field + ":\"" + message + "\""

	query = bleve.NewMatchPhraseQuery(queryString)
	return query
}

func (s *Shard) SearchTypePrefix(searchType interface{}) (query *query.PrefixQuery) {
	m := searchType.(map[string]interface{})
	var field, message string
	for key, value := range m {
		field = key
		message = value.(string)
	}
	queryString := field + ":\"" + message + "\""

	query = bleve.NewPrefixQuery(strings.ToLower(queryString))
	return query
}

func (s *Shard) SearchTypeFuzzy(searchType interface{}) (query *query.FuzzyQuery) {
	m := searchType.(map[string]interface{})
	var field, message string
	for key, value := range m {
		field = key
		message = value.(string)
	}
	queryString := field + ":\"" + message + "\""

	query = bleve.NewFuzzyQuery(strings.ToLower(queryString))
	return query
}

func (s *Shard) SearchTypeNumericRange(searchType interface{}) (query *query.ConjunctionQuery) {
	m := searchType.(map[string]interface{})
	var field string
	var message map[string]interface{}
	var gt, lt float64
	var v1incl, v2incl bool
	for key, value := range m {
		field = key
		message = value.(map[string]interface{})
	}
	for key, value := range message {
		if key == "gt" {
			gt = value.(float64)
			v1incl = false
		} else if key == "gte" {
			gt = value.(float64)
			v1incl = true
		} else if key == "lt" {
			lt = value.(float64)
			v2incl = false
		} else if key == "lte" {
			lt = value.(float64)
			v2incl = true
		}
	}
	gtString := field + ":"
	if v1incl {
		gtString += ">=" + fmt.Sprintf("%f", gt)
	} else {
		gtString += ">" + fmt.Sprintf("%f", gt)
	}
	ltString := field + ":"
	if v2incl {
		ltString += "<=" + fmt.Sprintf("%f", lt)
	} else {
		ltString += "<" + fmt.Sprintf("%f", lt)
	}
	gtQuery := bleve.NewQueryStringQuery(gtString)
	ltQuery := bleve.NewQueryStringQuery(ltString)
	query = bleve.NewConjunctionQuery(gtQuery, ltQuery)
	return query
}

func (s *Shard) SearchTypeBool(searchType interface{}) (query *query.BooleanQuery) {
	m := searchType.(map[string]interface{})
	query = bleve.NewBooleanQuery()
	var mustSearch, mustNotSearch, shouldSearch map[string]interface{}
	for key, value := range m {
		if key == "must" {
			mustSearch = value.(map[string]interface{})
		} else if key == "must_not" {
			mustNotSearch = value.(map[string]interface{})
		} else if key == "should" {
			shouldSearch = value.(map[string]interface{})
		}
	}
	for key, value := range mustSearch {
		if key == "match" {
			query.AddMust(s.SearchTypeMatch(value))
		} else if key == "match_phrase" {
			query.AddMust(s.SearchTypeMatchPhrase(value))
		} else if key == "prefix" {
			query.AddMust(s.SearchTypePrefix(value))
		} else if key == "fuzzy" {
			query.AddMust(s.SearchTypeFuzzy(value))
		} else if key == "range" {
			query.AddMust(s.SearchTypeNumericRange(value))
		}
	}
	for key, value := range mustNotSearch {
		if key == "match" {
			query.AddMustNot(s.SearchTypeMatch(value))
		} else if key == "match_phrase" {
			query.AddMustNot(s.SearchTypeMatchPhrase(value))
		} else if key == "prefix" {
			query.AddMustNot(s.SearchTypePrefix(value))
		} else if key == "fuzzy" {
			query.AddMustNot(s.SearchTypeFuzzy(value))
		} else if key == "range" {
			query.AddMustNot(s.SearchTypeNumericRange(value))
		}
	}
	for key, value := range shouldSearch {
		if key == "match" {
			query.AddShould(s.SearchTypeMatch(value))
		} else if key == "match_phrase" {
			query.AddShould(s.SearchTypeMatchPhrase(value))
		} else if key == "prefix" {
			query.AddShould(s.SearchTypePrefix(value))
		} else if key == "fuzzy" {
			query.AddShould(s.SearchTypeFuzzy(value))
		} else if key == "range" {
			query.AddShould(s.SearchTypeNumericRange(value))
		}
	}
	return query
}
