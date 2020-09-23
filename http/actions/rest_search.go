package actions

import (
	"encoding/json"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/blevesearch/bleve"
	"github.com/nqd/flat"
)

type RestSearch struct {
	ClusterService *cluster.Service
	IndicesService *indices.Service
}

func (h *RestSearch) Handle(r *RestRequest, reply ResponseListener) {
	indexName := r.PathParams["index"]

	var body map[string]interface{}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		reply(RestResponse{
			StatusCode: 400,
			Body: map[string]interface{}{
				"err": err,
			},
		})
		return
	}

	var qType map[string]interface{}
	if v, found := body["query"]; found {
		qType = v.(map[string]interface{})
	}

	clusterState := h.ClusterService.State()
	index := clusterState.Metadata.Indices[indexName].Index
	uuid := index.Uuid

	indexService, _ := h.IndicesService.IndexService(uuid)
	indexShard, _ := indexService.Shard(0)

	if k, found := qType["match"]; found {
		query := indexShard.SearchTypeMatch(k)
		searchRequest := bleve.NewSearchRequest(query)
		searchResults, err := indexShard.Search(searchRequest)
		if err != nil {
			reply(RestResponse{
				StatusCode: 400,
				Body: map[string]interface{}{
					"err": err,
				},
			})
			return
		}

		var maxScore float64
		var docList []interface{}
		for _, hits := range searchResults.Hits {
			doc, _ := indexShard.Get(hits.ID)
			src, _ := flat.Unflatten(doc, nil)
			hitJson := map[string]interface{}{
				"_index":  indexName,
				"_type":   "_doc",
				"_id":     hits.ID,
				"_score":  hits.Score,
				"_source": src,
			}
			if maxScore < hits.Score {
				maxScore = hits.Score
			}
			docList = append(docList, hitJson)
		}

		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"took":      searchResults.Took.Microseconds(),
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      1,
					"successful": 1,
					"skipped":    0,
					"failed":     0,
				},
				"hits": map[string]interface{}{
					"total": map[string]interface{}{
						"value":    1,
						"relation": "eq",
					},
					"max_score": maxScore,
					"hits":      docList,
				},
			},
		})
	} else if k, found := qType["match_phrase"]; found {
		query := indexShard.SearchTypeMatchPhrase(k)
		searchRequest := bleve.NewSearchRequest(query)
		searchResults, err := indexShard.Search(searchRequest)
		if err != nil {
			reply(RestResponse{
				StatusCode: 400,
				Body: map[string]interface{}{
					"err": err,
				},
			})
			return
		}

		var maxScore float64
		var docList []interface{}
		for _, hits := range searchResults.Hits {
			doc, _ := indexShard.Get(hits.ID)
			src, _ := flat.Unflatten(doc, nil)
			hitJson := map[string]interface{}{
				"_index":  indexName,
				"_type":   "_doc",
				"_id":     hits.ID,
				"_score":  hits.Score,
				"_source": src,
			}
			if maxScore < hits.Score {
				maxScore = hits.Score
			}
			docList = append(docList, hitJson)
		}

		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"took":      searchResults.Took.Microseconds(),
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      1,
					"successful": 1,
					"skipped":    0,
					"failed":     0,
				},
				"hits": map[string]interface{}{
					"total": map[string]interface{}{
						"value":    1,
						"relation": "eq",
					},
					"max_score": maxScore,
					"hits":      docList,
				},
			},
		})
	} else if _, found := qType["match_all"]; found {
		query := bleve.NewMatchAllQuery()
		searchRequest := bleve.NewSearchRequest(query)
		searchResults, err := indexShard.Search(searchRequest)
		if err != nil {
			reply(RestResponse{
				StatusCode: 400,
				Body: map[string]interface{}{
					"err": err,
				},
			})
			return
		}

		var maxScore float64
		var docList []interface{}
		for _, hits := range searchResults.Hits {
			doc, _ := indexShard.Get(hits.ID)
			src, _ := flat.Unflatten(doc, nil)
			hitJson := map[string]interface{}{
				"_index":  indexName,
				"_type":   "_doc",
				"_id":     hits.ID,
				"_score":  hits.Score,
				"_source": src,
			}
			if maxScore < hits.Score {
				maxScore = hits.Score
			}
			docList = append(docList, hitJson)
		}

		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"took":      searchResults.Took.Microseconds(),
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      1,
					"successful": 1,
					"skipped":    0,
					"failed":     0,
				},
				"hits": map[string]interface{}{
					"total": map[string]interface{}{
						"value":    1,
						"relation": "eq",
					},
					"max_score": maxScore,
					"hits":      docList,
				},
			},
		})
	} else if k, found := qType["prefix"]; found {
		query := indexShard.SearchTypePrefix(k)
		searchRequest := bleve.NewSearchRequest(query)
		searchResults, err := indexShard.Search(searchRequest)
		if err != nil {
			reply(RestResponse{
				StatusCode: 400,
				Body: map[string]interface{}{
					"err": err,
				},
			})
			return
		}

		var maxScore float64
		var docList []interface{}
		for _, hits := range searchResults.Hits {
			doc, _ := indexShard.Get(hits.ID)
			src, _ := flat.Unflatten(doc, nil)
			hitJson := map[string]interface{}{
				"_index":  indexName,
				"_type":   "_doc",
				"_id":     hits.ID,
				"_score":  hits.Score,
				"_source": src,
			}
			if maxScore < hits.Score {
				maxScore = hits.Score
			}
			docList = append(docList, hitJson)
		}

		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"took":      searchResults.Took.Microseconds(),
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      1,
					"successful": 1,
					"skipped":    0,
					"failed":     0,
				},
				"hits": map[string]interface{}{
					"total": map[string]interface{}{
						"value":    1,
						"relation": "eq",
					},
					"max_score": maxScore,
					"hits":      docList,
				},
			},
		})
	} else if k, found := qType["fuzzy"]; found {
		query := indexShard.SearchTypeFuzzy(k)
		searchRequest := bleve.NewSearchRequest(query)
		searchResults, err := indexShard.Search(searchRequest)
		if err != nil {
			reply(RestResponse{
				StatusCode: 400,
				Body: map[string]interface{}{
					"err": err,
				},
			})
			return
		}

		var maxScore float64
		var docList []interface{}
		for _, hits := range searchResults.Hits {
			doc, _ := indexShard.Get(hits.ID)
			src, _ := flat.Unflatten(doc, nil)
			hitJson := map[string]interface{}{
				"_index":  indexName,
				"_type":   "_doc",
				"_id":     hits.ID,
				"_score":  hits.Score,
				"_source": src,
			}
			if maxScore < hits.Score {
				maxScore = hits.Score
			}
			docList = append(docList, hitJson)
		}

		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"took":      searchResults.Took.Microseconds(),
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      1,
					"successful": 1,
					"skipped":    0,
					"failed":     0,
				},
				"hits": map[string]interface{}{
					"total": map[string]interface{}{
						"value":    1,
						"relation": "eq",
					},
					"max_score": maxScore,
					"hits":      docList,
				},
			},
		})
	} else if k, found := qType["bool"]; found {
		query := indexShard.SearchTypeBool(k)
		searchRequest := bleve.NewSearchRequest(query)
		searchResults, err := indexShard.Search(searchRequest)
		if err != nil {
			reply(RestResponse{
				StatusCode: 400,
				Body: map[string]interface{}{
					"err": err,
				},
			})
			return
		}

		var maxScore float64
		var docList []interface{}
		for _, hits := range searchResults.Hits {
			doc, _ := indexShard.Get(hits.ID)
			src, _ := flat.Unflatten(doc, nil)
			hitJson := map[string]interface{}{
				"_index":  indexName,
				"_type":   "_doc",
				"_id":     hits.ID,
				"_score":  hits.Score,
				"_source": src,
			}
			if maxScore < hits.Score {
				maxScore = hits.Score
			}
			docList = append(docList, hitJson)
		}

		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"took":      searchResults.Took.Microseconds(),
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      1,
					"successful": 1,
					"skipped":    0,
					"failed":     0,
				},
				"hits": map[string]interface{}{
					"total": map[string]interface{}{
						"value":    1,
						"relation": "eq",
					},
					"max_score": maxScore,
					"hits":      docList,
				},
			},
		})
	} else if k, found := qType["range"]; found {
		query := indexShard.SearchTypeNumericRange(k)
		searchRequest := bleve.NewSearchRequest(query)
		searchResults, err := indexShard.Search(searchRequest)
		if err != nil {
			reply(RestResponse{
				StatusCode: 400,
				Body: map[string]interface{}{
					"err": err,
				},
			})
			return
		}

		var maxScore float64
		var docList []interface{}
		for _, hits := range searchResults.Hits {
			doc, _ := indexShard.Get(hits.ID)
			src, _ := flat.Unflatten(doc, nil)
			hitJson := map[string]interface{}{
				"_index":  indexName,
				"_type":   "_doc",
				"_id":     hits.ID,
				"_score":  hits.Score,
				"_source": src,
			}
			if maxScore < hits.Score {
				maxScore = hits.Score
			}
			docList = append(docList, hitJson)
		}

		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"took":      searchResults.Took.Microseconds(),
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      1,
					"successful": 1,
					"skipped":    0,
					"failed":     0,
				},
				"hits": map[string]interface{}{
					"total": map[string]interface{}{
						"value":    1,
						"relation": "eq",
					},
					"max_score": maxScore,
					"hits":      docList,
				},
			},
		})
	} else {
		reply(RestResponse{
			StatusCode: 400,
			Body: map[string]interface{}{
				"err": "search error",
			},
		})
		return
	}
}
