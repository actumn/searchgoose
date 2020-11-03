package index

import (
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"strings"
)

func SearchTypeMatch(searchType interface{}) *query.MatchQuery {
	m := searchType.(map[string]interface{})
	var field, message string
	for key, value := range m {
		field = key
		switch v := value.(type) {
		case string:
			message = v
		case map[string]interface{}:
			message = v["query"].(string)
		}
	}
	queryString := field + ":\"" + message + "\""

	return bleve.NewMatchQuery(queryString)
}

func SearchTypeMatchPhrase(searchType interface{}) *query.PhraseQuery {
	m := searchType.(map[string]interface{})
	var field, message string
	for key, value := range m {
		field = key
		switch v := value.(type) {
		case string:
			message = v
		case map[string]interface{}:
			message = v["query"].(string)
		}
	}

	return bleve.NewPhraseQuery(strings.Fields(strings.ToLower(message)), field)
}

func SearchTypePrefix(searchType interface{}) *query.PrefixQuery {
	m := searchType.(map[string]interface{})
	var _, message string
	for key, value := range m {
		_ = key
		switch v := value.(type) {
		case string:
			message = v
		case map[string]interface{}:
			message = v["query"].(string)
		}
	}

	return bleve.NewPrefixQuery(strings.ToLower(message))
}

func SearchTypeFuzzy(searchType interface{}) *query.FuzzyQuery {
	m := searchType.(map[string]interface{})
	var _, message string
	for key, value := range m {
		_ = key
		message = value.(string)
	}

	return bleve.NewFuzzyQuery(strings.ToLower(message))
}

func SearchTypeNumericRange(searchType interface{}) *query.ConjunctionQuery {
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

func SearchTypeSimpleQueryString(searchType interface{}) *query.QueryStringQuery {
	m := searchType.(map[string]interface{})
	var message string
	if k, found := m["query"]; found {
		message = k.(string)
	}

	return bleve.NewQueryStringQuery(message)
}

func SearchTypeBool(searchType interface{}) *query.BooleanQuery {
	m := searchType.(map[string]interface{})
	searchQuery := bleve.NewBooleanQuery()
	var mustSearch, mustNotSearch, shouldSearch map[string]interface{}
	for key, value := range m {
		switch key {
		case "must", "filter":
			queryMap, ok := value.([]interface{})
			if ok {
				mustSearch = queryMap[0].(map[string]interface{})
			} else {
				mustSearch = value.(map[string]interface{})
			}
		case "must_not":
			queryMap, ok := value.([]interface{})
			if ok {
				mustNotSearch = queryMap[0].(map[string]interface{})
			} else {
				mustNotSearch = value.(map[string]interface{})
			}
		case "should":
			queryMap, ok := value.([]interface{})
			if ok {
				shouldSearch = queryMap[0].(map[string]interface{})
			} else {
				shouldSearch = value.(map[string]interface{})
			}
		}
	}
	for key, value := range mustSearch {
		switch key {
		case "match", "term":
			searchQuery.AddMust(SearchTypeMatch(value))
		case "match_phrase":
			searchQuery.AddMust(SearchTypeMatchPhrase(value))
		case "prefix":
			searchQuery.AddMust(SearchTypePrefix(value))
		case "fuzzy":
			searchQuery.AddMust(SearchTypeFuzzy(value))
		case "range":
			searchQuery.AddMust(SearchTypeNumericRange(value))
		case "bool":
			searchQuery.AddMust(SearchTypeBool(value))
		case "simple_query_string":
			searchQuery.AddMust(SearchTypeSimpleQueryString(value))
		}
	}
	for key, value := range mustNotSearch {
		switch key {
		case "match", "term":
			searchQuery.AddMustNot(SearchTypeMatch(value))
		case "match_phrase":
			searchQuery.AddMustNot(SearchTypeMatchPhrase(value))
		case "prefix":
			searchQuery.AddMustNot(SearchTypePrefix(value))
		case "fuzzy":
			searchQuery.AddMustNot(SearchTypeFuzzy(value))
		case "range":
			searchQuery.AddMustNot(SearchTypeNumericRange(value))
		case "bool":
			searchQuery.AddMustNot(SearchTypeBool(value))
		case "simple_query_string":
			searchQuery.AddMustNot(SearchTypeSimpleQueryString(value))
		}
	}
	for key, value := range shouldSearch {
		switch key {
		case "match", "term":
			searchQuery.AddShould(SearchTypeMatch(value))
		case "match_phrase":
			searchQuery.AddShould(SearchTypeMatchPhrase(value))
		case "prefix":
			searchQuery.AddShould(SearchTypePrefix(value))
		case "fuzzy":
			searchQuery.AddShould(SearchTypeFuzzy(value))
		case "range":
			searchQuery.AddShould(SearchTypeNumericRange(value))
		case "bool":
			searchQuery.AddShould(SearchTypeBool(value))
		case "simple_query_string":
			searchQuery.AddShould(SearchTypeSimpleQueryString(value))
		}
	}
	return searchQuery
}
