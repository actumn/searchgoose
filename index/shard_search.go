package index

import (
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"strings"
)

func (s *Shard) SearchTypeMatch(searchType interface{}) *query.MatchQuery {
	m := searchType.(map[string]interface{})
	var field, message string
	for key, value := range m {
		field = key
		message = value.(string)
	}
	queryString := field + ":\"" + message + "\""

	return bleve.NewMatchQuery(queryString)
}

func (s *Shard) SearchTypeMatchPhrase(searchType interface{}) *query.MatchPhraseQuery {
	m := searchType.(map[string]interface{})
	var field, message string
	for key, value := range m {
		field = key
		message = value.(string)
	}
	queryString := field + ":\"" + message + "\""

	return bleve.NewMatchPhraseQuery(queryString)
}

func (s *Shard) SearchTypePrefix(searchType interface{}) *query.PrefixQuery {
	m := searchType.(map[string]interface{})
	var field, message string
	for key, value := range m {
		field = key
		message = value.(string)
	}
	queryString := field + ":\"" + message + "\""

	return bleve.NewPrefixQuery(strings.ToLower(queryString))
}

func (s *Shard) SearchTypeFuzzy(searchType interface{}) *query.FuzzyQuery {
	m := searchType.(map[string]interface{})
	var field, message string
	for key, value := range m {
		field = key
		message = value.(string)
	}
	queryString := field + ":\"" + message + "\""

	return bleve.NewFuzzyQuery(strings.ToLower(queryString))
}

func (s *Shard) SearchTypeNumericRange(searchType interface{}) *query.ConjunctionQuery {
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
	return bleve.NewConjunctionQuery(gtQuery, ltQuery)
}

func (s *Shard) SearchTypeBool(searchType interface{}) *query.BooleanQuery {
	m := searchType.(map[string]interface{})
	searchQuery := bleve.NewBooleanQuery()
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
			searchQuery.AddMust(s.SearchTypeMatch(value))
		} else if key == "match_phrase" {
			searchQuery.AddMust(s.SearchTypeMatchPhrase(value))
		} else if key == "prefix" {
			searchQuery.AddMust(s.SearchTypePrefix(value))
		} else if key == "fuzzy" {
			searchQuery.AddMust(s.SearchTypeFuzzy(value))
		} else if key == "range" {
			searchQuery.AddMust(s.SearchTypeNumericRange(value))
		}
	}
	for key, value := range mustNotSearch {
		if key == "match" {
			searchQuery.AddMustNot(s.SearchTypeMatch(value))
		} else if key == "match_phrase" {
			searchQuery.AddMustNot(s.SearchTypeMatchPhrase(value))
		} else if key == "prefix" {
			searchQuery.AddMustNot(s.SearchTypePrefix(value))
		} else if key == "fuzzy" {
			searchQuery.AddMustNot(s.SearchTypeFuzzy(value))
		} else if key == "range" {
			searchQuery.AddMustNot(s.SearchTypeNumericRange(value))
		}
	}
	for key, value := range shouldSearch {
		if key == "match" {
			searchQuery.AddShould(s.SearchTypeMatch(value))
		} else if key == "match_phrase" {
			searchQuery.AddShould(s.SearchTypeMatchPhrase(value))
		} else if key == "prefix" {
			searchQuery.AddShould(s.SearchTypePrefix(value))
		} else if key == "fuzzy" {
			searchQuery.AddShould(s.SearchTypeFuzzy(value))
		} else if key == "range" {
			searchQuery.AddShould(s.SearchTypeNumericRange(value))
		}
	}
	return searchQuery
}