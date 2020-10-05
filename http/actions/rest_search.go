package actions

import (
	"encoding/json"
	"github.com/actumn/searchgoose/index"
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
	idx := clusterState.Metadata.Indices[indexName].Index
	shardNum := clusterState.Metadata.Indices[indexName].RoutingNumShards
	uuid := idx.Uuid

	indexService, _ := h.IndicesService.IndexService(uuid)

	var shards []*index.Shard
	for i := 0; i < shardNum; i++ {
		indexShard, _ := indexService.Shard(i)
		shards = append(shards, indexShard)
	}

	if k, found := qType["match"]; found {
		var results []*bleve.SearchResult
		query := index.SearchTypeMatch(k)
		searchRequest := bleve.NewSearchRequest(query)
		for _, s := range shards {
			res, err := s.Search(searchRequest)
			if err != nil {
				reply(RestResponse{
					StatusCode: 400,
					Body: map[string]interface{}{
						"err": err,
					},
				})
				return
			}
			results = append(results, res)
		}

		var maxScore float64
		var docList []interface{}
		for i, s := range shards {
			for _, hits := range results[i].Hits {
				doc, _ := s.Get(hits.ID)
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
		}
		var took int64
		for _, r := range results {
			took += r.Took.Microseconds()
		}

		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"took":      took,
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      shardNum,
					"successful": shardNum,
					"skipped":    0,
					"failed":     0,
				},
				"hits": map[string]interface{}{
					"total": map[string]interface{}{
						"value":    len(docList),
						"relation": "eq",
					},
					"max_score": maxScore,
					"hits":      docList,
				},
			},
		})
	} else if k, found := qType["match_phrase"]; found {
		var results []*bleve.SearchResult
		query := index.SearchTypeMatchPhrase(k)
		searchRequest := bleve.NewSearchRequest(query)
		for _, s := range shards {
			res, err := s.Search(searchRequest)
			if err != nil {
				reply(RestResponse{
					StatusCode: 400,
					Body: map[string]interface{}{
						"err": err,
					},
				})
				return
			}
			results = append(results, res)
		}

		var maxScore float64
		var docList []interface{}
		for i, s := range shards {
			for _, hits := range results[i].Hits {
				doc, _ := s.Get(hits.ID)
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
		}
		var took int64
		for _, r := range results {
			took += r.Took.Microseconds()
		}

		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"took":      took,
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      shardNum,
					"successful": shardNum,
					"skipped":    0,
					"failed":     0,
				},
				"hits": map[string]interface{}{
					"total": map[string]interface{}{
						"value":    len(docList),
						"relation": "eq",
					},
					"max_score": maxScore,
					"hits":      docList,
				},
			},
		})
	} else if _, found := qType["match_all"]; found {
		var results []*bleve.SearchResult
		query := bleve.NewMatchAllQuery()
		searchRequest := bleve.NewSearchRequest(query)
		for _, s := range shards {
			res, err := s.Search(searchRequest)
			if err != nil {
				reply(RestResponse{
					StatusCode: 400,
					Body: map[string]interface{}{
						"err": err,
					},
				})
				return
			}
			results = append(results, res)
		}

		var maxScore float64
		var docList []interface{}
		for i, s := range shards {
			for _, hits := range results[i].Hits {
				doc, _ := s.Get(hits.ID)
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
		}
		var took int64
		for _, r := range results {
			took += r.Took.Microseconds()
		}

		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"took":      took,
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      shardNum,
					"successful": shardNum,
					"skipped":    0,
					"failed":     0,
				},
				"hits": map[string]interface{}{
					"total": map[string]interface{}{
						"value":    len(docList),
						"relation": "eq",
					},
					"max_score": maxScore,
					"hits":      docList,
				},
			},
		})
	} else if k, found := qType["prefix"]; found {
		var results []*bleve.SearchResult
		query := index.SearchTypePrefix(k)
		searchRequest := bleve.NewSearchRequest(query)
		for _, s := range shards {
			res, err := s.Search(searchRequest)
			if err != nil {
				reply(RestResponse{
					StatusCode: 400,
					Body: map[string]interface{}{
						"err": err,
					},
				})
				return
			}
			results = append(results, res)
		}

		var maxScore float64
		var docList []interface{}
		for i, s := range shards {
			for _, hits := range results[i].Hits {
				doc, _ := s.Get(hits.ID)
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
		}
		var took int64
		for _, r := range results {
			took += r.Took.Microseconds()
		}

		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"took":      took,
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      shardNum,
					"successful": shardNum,
					"skipped":    0,
					"failed":     0,
				},
				"hits": map[string]interface{}{
					"total": map[string]interface{}{
						"value":    len(docList),
						"relation": "eq",
					},
					"max_score": maxScore,
					"hits":      docList,
				},
			},
		})
	} else if k, found := qType["fuzzy"]; found {
		var results []*bleve.SearchResult
		query := index.SearchTypeFuzzy(k)
		searchRequest := bleve.NewSearchRequest(query)
		for _, s := range shards {
			res, err := s.Search(searchRequest)
			if err != nil {
				reply(RestResponse{
					StatusCode: 400,
					Body: map[string]interface{}{
						"err": err,
					},
				})
				return
			}
			results = append(results, res)
		}

		var maxScore float64
		var docList []interface{}
		for i, s := range shards {
			for _, hits := range results[i].Hits {
				doc, _ := s.Get(hits.ID)
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
		}
		var took int64
		for _, r := range results {
			took += r.Took.Microseconds()
		}

		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"took":      took,
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      shardNum,
					"successful": shardNum,
					"skipped":    0,
					"failed":     0,
				},
				"hits": map[string]interface{}{
					"total": map[string]interface{}{
						"value":    len(docList),
						"relation": "eq",
					},
					"max_score": maxScore,
					"hits":      docList,
				},
			},
		})
	} else if k, found := qType["bool"]; found {
		var results []*bleve.SearchResult
		query := index.SearchTypeBool(k)
		searchRequest := bleve.NewSearchRequest(query)
		for _, s := range shards {
			res, err := s.Search(searchRequest)
			if err != nil {
				reply(RestResponse{
					StatusCode: 400,
					Body: map[string]interface{}{
						"err": err,
					},
				})
				return
			}
			results = append(results, res)
		}

		var maxScore float64
		var docList []interface{}
		for i, s := range shards {
			for _, hits := range results[i].Hits {
				doc, _ := s.Get(hits.ID)
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
		}
		var took int64
		for _, r := range results {
			took += r.Took.Microseconds()
		}

		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"took":      took,
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      shardNum,
					"successful": shardNum,
					"skipped":    0,
					"failed":     0,
				},
				"hits": map[string]interface{}{
					"total": map[string]interface{}{
						"value":    len(docList),
						"relation": "eq",
					},
					"max_score": maxScore,
					"hits":      docList,
				},
			},
		})
	} else if k, found := qType["range"]; found {
		var results []*bleve.SearchResult
		query := index.SearchTypeNumericRange(k)
		searchRequest := bleve.NewSearchRequest(query)
		for _, s := range shards {
			res, err := s.Search(searchRequest)
			if err != nil {
				reply(RestResponse{
					StatusCode: 400,
					Body: map[string]interface{}{
						"err": err,
					},
				})
				return
			}
			results = append(results, res)
		}

		var maxScore float64
		var docList []interface{}
		for i, s := range shards {
			for _, hits := range results[i].Hits {
				doc, _ := s.Get(hits.ID)
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
		}
		var took int64
		for _, r := range results {
			took += r.Took.Microseconds()
		}

		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"took":      took,
				"timed_out": false,
				"_shards": map[string]interface{}{
					"total":      shardNum,
					"successful": shardNum,
					"skipped":    0,
					"failed":     0,
				},
				"hits": map[string]interface{}{
					"total": map[string]interface{}{
						"value":    len(docList),
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
