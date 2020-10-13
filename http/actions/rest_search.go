package actions

import (
	"encoding/json"
	"github.com/actumn/searchgoose/index"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/nqd/flat"
	"strings"
)

const (
	SearchAction = "search"
)

type RestSearch struct {
	clusterService   *cluster.Service
	indicesService   *indices.Service
	transportService *transport.Service
}

type SearchRequest struct {
	searchType string
	searchBody map[string]interface{}
}

func (r *SearchRequest) ToBytes() []byte {
	gap := 12 - len(r.searchType)
	t := []byte(r.searchType + strings.Repeat("=", gap))
	b, _ := json.Marshal(r.searchBody)

	return append(t, b...)
}

func SearchRequestFromBytes(b []byte) *SearchRequest {
	t := string(b[0:12])
	var body map[string]interface{}
	_ = json.Unmarshal(b[12:] ,&body)

	return &SearchRequest{
		searchType: t,
		searchBody: body,
	}
}

type SearchResponse struct {

}

func (r *SearchResponse) ToBytes() []byte {

}

func SearchResponseFromBytes(b []byte) *SearchResponse {

}

func NewRestSearch(clusterService *cluster.Service, indicesService *indices.Service, transportService *transport.Service) *RestSearch {
	transportService.RegisterRequestHandler(SearchAction, func(channel transport.ReplyChannel, req []byte) []byte {
		request := SearchRequestFromBytes(req)

		res := SearchResponse{}
		return res.ToBytes()
	})

	for i := 0; i < shardCount; i++ {

		request := SearchRequest{

		}
		transportService.SendRequest(node, SearchAction, request.ToBytes(), func(response []byte) {
			res := SearchResponseFromBytes(response)

		})

	}

	return &RestSearch{
		clusterService:   clusterService,
		indicesService:   indicesService,
		transportService: transportService,
	}
}

type SearchResultData struct {
	results  []*bleve.SearchResult
	docList  []interface{}
	maxScore float64
	took     int64
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

	clusterState := h.clusterService.State()
	idx := clusterState.Metadata.Indices[indexName].Index
	shardNum := clusterState.Metadata.Indices[indexName].RoutingNumShards
	uuid := idx.Uuid

	indexService, _ := h.indicesService.IndexService(uuid)

	var shards []*index.Shard
	for i := 0; i < shardNum; i++ {
		indexShard, _ := indexService.Shard(i)
		shards = append(shards, indexShard)
	}

	var q query.Query
	var data SearchResultData

	if k, found := qType["match"]; found {
		q = index.SearchTypeMatch(k)
	} else if k, found := qType["match_phrase"]; found {
		q = index.SearchTypeMatchPhrase(k)
	} else if _, found := qType["match_all"]; found {
		q = bleve.NewMatchAllQuery()
	} else if k, found := qType["prefix"]; found {
		q = index.SearchTypePrefix(k)
	} else if k, found := qType["fuzzy"]; found {
		q = index.SearchTypeFuzzy(k)
	} else if k, found := qType["bool"]; found {
		q = index.SearchTypeBool(k)
	} else if k, found := qType["range"]; found {
		q = index.SearchTypeNumericRange(k)
	} else {
		reply(RestResponse{
			StatusCode: 400,
			Body: map[string]interface{}{
				"err": "search error",
			},
		})
		return
	}

	for shardNum, shardRouting := range clusterState.RoutingTable.IndicesRouting[indexName].Shards {
		node := clusterState.Nodes.Nodes[shardRouting.Primary.CurrentNodeId]

		h.transportService.SendRequest(*node, SearchAction, []byte("data"), func(response []byte) {

		})
	}

	searchRequest := bleve.NewSearchRequest(q)
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
		data.results = append(data.results, res)
	}

	for i, s := range shards {
		for _, hits := range data.results[i].Hits {
			doc, _ := s.Get(hits.ID)
			src, _ := flat.Unflatten(doc, nil)
			hitJson := map[string]interface{}{
				"_index":  indexName,
				"_type":   "_doc",
				"_id":     hits.ID,
				"_score":  hits.Score,
				"_source": src,
			}
			if data.maxScore < hits.Score {
				data.maxScore = hits.Score
			}
			data.docList = append(data.docList, hitJson)
		}
	}
	for _, r := range data.results {
		data.took += r.Took.Microseconds()
	}

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"took":      data.took,
			"timed_out": false,
			"_shards": map[string]interface{}{
				"total":      shardNum,
				"successful": shardNum,
				"skipped":    0,
				"failed":     0,
			},
			"hits": map[string]interface{}{
				"total": map[string]interface{}{
					"value":    len(data.docList),
					"relation": "eq",
				},
				"max_score": data.maxScore,
				"hits":      data.docList,
			},
		},
	})
}
