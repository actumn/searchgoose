package actions

import (
	"bytes"
	"encoding/json"
	"github.com/actumn/searchgoose/index"
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/nqd/flat"
	"github.com/sirupsen/logrus"
)

const (
	SearchAction = "search"
)

type RestSearch struct {
	clusterService              *cluster.Service
	indicesService              *indices.Service
	indexNameExpressionResolver *indices.NameExpressionResolver
	transportService            *transport.Service
}

type SearchRequest struct {
	SearchIndex string
	ShardId     state.ShardId
	SearchBody  map[string]interface{}
}

func (r *SearchRequest) toBytes() []byte {
	var buffer bytes.Buffer
	enc := json.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func SearchRequestFromBytes(b []byte) *SearchRequest {
	buffer := bytes.NewBuffer(b)
	decoder := json.NewDecoder(buffer)
	var req SearchRequest
	if err := decoder.Decode(&req); err != nil {
		logrus.Fatal(err)
	}
	return &req
}

type SearchResultData struct {
	Results  *bleve.SearchResult
	DocList  []interface{}
	MaxScore float64
	Took     int64
}

type SearchResponse struct {
	SearchResult SearchResultData
}

func (r *SearchResponse) ToBytes() []byte {
	var buffer bytes.Buffer
	enc := json.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func SearchResponseFromBytes(b []byte) *SearchResponse {
	buffer := bytes.NewBuffer(b)
	decoder := json.NewDecoder(buffer)
	var req SearchResponse
	if err := decoder.Decode(&req); err != nil {
		logrus.Fatal(err)
	}
	return &req
}

func NewRestSearch(clusterService *cluster.Service, indicesService *indices.Service, indexNameExpressionResolver *indices.NameExpressionResolver, transportService *transport.Service) *RestSearch {
	transportService.RegisterRequestHandler(SearchAction, func(channel transport.ReplyChannel, req []byte) {
		request := SearchRequestFromBytes(req)
		indexName := request.SearchIndex
		body := request.SearchBody

		var qType map[string]interface{}
		if v, found := body["query"]; found {
			qType = v.(map[string]interface{})
		}

		indexService, _ := indicesService.IndexService(request.ShardId.Index.Uuid)
		indexShard, _ := indexService.Shard(request.ShardId.ShardId)

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
			logrus.Fatal("search error")
			return
		}

		searchRequest := bleve.NewSearchRequest(q)
		r, err := indexShard.Search(searchRequest)
		if err != nil {
			logrus.Fatal(err)
		}
		data.Results = r

		for _, hits := range data.Results.Hits {
			doc, _ := indexShard.Get(hits.ID)
			src, _ := flat.Unflatten(doc, nil)
			hitJson := map[string]interface{}{
				"_index":  indexName,
				"_type":   "_doc",
				"_id":     hits.ID,
				"_score":  hits.Score,
				"_source": src,
			}
			if data.MaxScore < hits.Score {
				data.MaxScore = hits.Score
			}
			data.DocList = append(data.DocList, hitJson)
		}
		data.Took += data.Results.Took.Microseconds()

		res := SearchResponse{data}
		channel.SendMessage("", res.ToBytes())
	})

	return &RestSearch{
		clusterService:              clusterService,
		indicesService:              indicesService,
		indexNameExpressionResolver: indexNameExpressionResolver,
		transportService:            transportService,
	}
}

func (h *RestSearch) Handle(r *RestRequest, reply ResponseListener) {
	indexExpression := r.PathParams["index"]

	var body map[string]interface{}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		logrus.Fatal(err)
	}
	clusterState := h.clusterService.State()
	indexName := h.indexNameExpressionResolver.ConcreteSingleIndex(*clusterState, indexExpression).Name
	shardNum := clusterState.Metadata.Indices[indexName].RoutingNumShards

	totalResults := make(chan SearchResultData, shardNum)

	for _, shardRouting := range clusterState.RoutingTable.IndicesRouting[indexName].Shards {
		req := SearchRequest{
			SearchIndex: indexName,
			ShardId:     shardRouting.ShardId,
			SearchBody:  body,
		}
		h.transportService.SendRequest(*clusterState.Nodes.Nodes[shardRouting.Primary.CurrentNodeId], SearchAction, req.toBytes(), func(response []byte) {
			totalResults <- SearchResponseFromBytes(response).SearchResult
		})
	}

	var data struct {
		DocList  []interface{}
		MaxScore float64
		Took     int64
	}
	for i := 0; i < shardNum; i++ {
		d := <-totalResults
		data.Took += d.Took
		if d.DocList != nil {
			data.DocList = append(data.DocList, d.DocList)
		}
		if data.MaxScore <= d.MaxScore {
			data.MaxScore = d.MaxScore
		}
	}

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"took":      data.Took,
			"timed_out": false,
			"_shards": map[string]interface{}{
				"total":      shardNum,
				"successful": shardNum,
				"skipped":    0,
				"failed":     0,
			},
			"hits": map[string]interface{}{
				"total": map[string]interface{}{
					"value":    len(data.DocList),
					"relation": "eq",
				},
				"max_score": data.MaxScore,
				"hits":      data.DocList,
			},
		},
	})
}
