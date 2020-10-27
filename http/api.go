package http

import (
	"bytes"
	"encoding/json"
	"github.com/actumn/searchgoose/http/actions"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

func requestFromCtx(ctx *fasthttp.RequestCtx) actions.RestRequest {
	request := actions.RestRequest{
		Path:        string(ctx.Path()),
		Header:      map[string][]byte{},
		QueryParams: map[string][]byte{},
		Body:        ctx.Request.Body(),
	}
	if bytes.Compare(ctx.Method(), []byte("GET")) == 0 {
		request.Method = actions.GET
	} else if bytes.Compare(ctx.Method(), []byte("POST")) == 0 {
		request.Method = actions.POST
	} else if bytes.Compare(ctx.Method(), []byte("PUT")) == 0 {
		request.Method = actions.PUT
	} else if bytes.Compare(ctx.Method(), []byte("DELETE")) == 0 {
		request.Method = actions.DELETE
	} else if bytes.Compare(ctx.Method(), []byte("HEAD")) == 0 {
		request.Method = actions.HEAD
	}
	ctx.Request.Header.VisitAll(func(key, value []byte) {
		request.Header[string(key)] = value
	})
	ctx.QueryArgs().VisitAll(func(key, value []byte) {
		request.QueryParams[string(key)] = value
	})

	return request
}

type RequestController struct {
	pathTrie *pathTrie
}

func (c *RequestController) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	request := requestFromCtx(ctx)
	if request.Path == "/favicon.ico" {
		return
	}
	logrus.Info(string(ctx.Method()), " ", string(ctx.Request.RequestURI()), " ", string(ctx.Request.Body()))
	allHandlers := c.pathTrie.retrieveAll(request.Path)
	for {
		h, params, err := allHandlers()
		request.PathParams = params
		if err != nil {
			logrus.Warn(string(ctx.Method()), " ", request.Path, " ", err)
			ctx.Response.Header.SetCanonical([]byte("Content-Type"), []byte("application/json"))
			ctx.Response.SetStatusCode(400)
			if err := json.NewEncoder(ctx).Encode(map[string]string{
				"msg": "no route",
			}); err != nil {
				logrus.Error(err)
			}
			return
		}
		methodHandlers, ok := h.(actions.MethodHandlers)
		if !ok {
			continue
		}

		handler, ok := methodHandlers[request.Method]
		if !ok {
			continue
		}

		handler.Handle(&request, func(response actions.RestResponse) {
			logrus.Debug("reply on ", string(ctx.Method()), " ", request.Path)
			ctx.Response.SetStatusCode(response.StatusCode)
			ctx.Response.Header.SetCanonical([]byte("Content-Type"), []byte("application/json"))
			if err := json.NewEncoder(ctx).Encode(response.Body); err != nil {
				logrus.Error(err)
			}
		})
		return
	}
}

type Bootstrap struct {
	s *fasthttp.Server
}

func New(
	clusterService *cluster.Service,
	clusterMetadataCreateIndexService *cluster.MetadataCreateIndexService,
	clusterMetadataDeleteIndexService *cluster.MetadataDeleteIndexService,
	clusterMetadataIndexAliasService *cluster.MetadataIndexAliasService,
	indicesService *indices.Service,
	transportService *transport.Service,
	indexNameExpressionResolver *indices.NameExpressionResolver,
) *Bootstrap {
	c := RequestController{}
	c.pathTrie = newPathTrie()
	c.pathTrie.insert("/", actions.MethodHandlers{
		actions.GET: actions.NewRestMain(clusterService),
	})
	c.pathTrie.insert("/_aliases", actions.MethodHandlers{
		actions.POST: actions.NewRestPostIndexAlias(clusterMetadataIndexAliasService),
	})
	c.pathTrie.insert("/_alias/{name}", actions.MethodHandlers{
		actions.GET: actions.NewRestGetIndexAlias(clusterService, indexNameExpressionResolver),
	})
	c.pathTrie.insert("/_xpack", actions.MethodHandlers{
		actions.GET: &actions.RestXpack{},
	})

	///////////////////////////// nodes ///////////////////////////////////
	nodeInfosAction := actions.NewRestNodesInfo(clusterService, transportService)
	c.pathTrie.insert("/_nodes", actions.MethodHandlers{
		actions.GET: nodeInfosAction,
	})
	c.pathTrie.insert("/_nodes/{nodeId}", actions.MethodHandlers{
		actions.GET: nodeInfosAction,
	})
	c.pathTrie.insert("/_nodes/{nodeId}/_all", actions.MethodHandlers{
		actions.GET: nodeInfosAction,
	})

	nodeStatsAction := actions.NewRestNodesStats(clusterService, transportService)
	c.pathTrie.insert("/_nodes/{nodeId}/stats", actions.MethodHandlers{
		actions.GET: nodeStatsAction,
	})
	c.pathTrie.insert("/_nodes/stats", actions.MethodHandlers{
		actions.GET: nodeStatsAction,
	})
	c.pathTrie.insert("/_nodes/stats/{metric}", actions.MethodHandlers{
		actions.GET: nodeStatsAction,
	})

	///////////////////////////// cat /////////////////////////////////////
	c.pathTrie.insert("/_cat/templates", actions.MethodHandlers{
		actions.GET: &actions.RestCatTemplates{},
	})
	c.pathTrie.insert("/_cat/templates/{name}", actions.MethodHandlers{
		actions.GET: &actions.RestCatTemplates{},
	})
	c.pathTrie.insert("/_cat/nodes", actions.MethodHandlers{
		actions.GET: actions.NewRestCatNodes(clusterService, transportService),
	})
	c.pathTrie.insert("/_cat/indices", actions.MethodHandlers{
		actions.GET: actions.NewRestCatIndices(clusterService, indexNameExpressionResolver, transportService),
	})

	//////////////////////////// cluster //////////////////////////////////
	c.pathTrie.insert("/_cluster/health", actions.MethodHandlers{
		actions.GET: actions.NewRestClusterHealth(clusterService, indexNameExpressionResolver),
	})
	c.pathTrie.insert("/_cluster/state/metadata", actions.MethodHandlers{
		actions.GET: actions.NewRestClusterStateMetadata(clusterService),
	})
	c.pathTrie.insert("/_cluster/stats", actions.MethodHandlers{
		actions.GET: actions.NewRestClusterStats(clusterService, transportService, indicesService),
	})

	//////////////////////////// index ////////////////////////////////////
	c.pathTrie.insert("/{index}", actions.MethodHandlers{
		actions.GET:    actions.NewRestGetIndex(clusterService, indexNameExpressionResolver),
		actions.PUT:    actions.NewRestPutIndex(clusterMetadataCreateIndexService),
		actions.DELETE: actions.NewRestDeleteIndex(clusterService, indexNameExpressionResolver, clusterMetadataDeleteIndexService),
		actions.HEAD:   actions.NewRestHeadIndex(clusterService, indexNameExpressionResolver),
	})
	c.pathTrie.insert("/{index}/_doc", actions.MethodHandlers{
		actions.POST: actions.NewRestIndexDoc(clusterService, clusterMetadataCreateIndexService, indicesService, indexNameExpressionResolver, transportService),
	})
	c.pathTrie.insert("/{index}/_doc/{id}", actions.MethodHandlers{
		actions.GET:    actions.NewRestGetDoc(clusterService, indicesService, indexNameExpressionResolver, transportService),
		actions.PUT:    actions.NewRestIndexDocId(clusterService, clusterMetadataCreateIndexService, indicesService, indexNameExpressionResolver, transportService),
		actions.DELETE: actions.NewRestDeleteDoc(clusterService, indicesService, indexNameExpressionResolver, transportService),
	})
	c.pathTrie.insert("/{index}/{type}/{id}", actions.MethodHandlers{ // deprecated but just for elasticsearch-HQ
		actions.GET:    actions.NewRestGetDoc(clusterService, indicesService, indexNameExpressionResolver, transportService),
		actions.PUT:    actions.NewRestIndexDocId(clusterService, clusterMetadataCreateIndexService, indicesService, indexNameExpressionResolver, transportService),
		actions.DELETE: actions.NewRestDeleteDoc(clusterService, indicesService, indexNameExpressionResolver, transportService),
	})
	c.pathTrie.insert("/{index}/_search", actions.MethodHandlers{
		actions.GET:  actions.NewRestSearch(clusterService, indicesService, indexNameExpressionResolver, transportService),
		actions.POST: actions.NewRestSearch(clusterService, indicesService, indexNameExpressionResolver, transportService),
	})
	c.pathTrie.insert("/{index}/_refresh", actions.MethodHandlers{
		actions.GET:  actions.NewRestRefresh(clusterService),
		actions.POST: actions.NewRestRefresh(clusterService),
	})

	indicesStatsAction := actions.NewRestIndicesStatsAction(clusterService, indicesService, indexNameExpressionResolver, transportService)
	c.pathTrie.insert("/_stats", actions.MethodHandlers{
		actions.GET: indicesStatsAction,
	})
	c.pathTrie.insert("/{index}/_stats", actions.MethodHandlers{
		actions.GET: indicesStatsAction,
	})
	//c.pathTrie.insert("/{index}/_bulk", actions.MethodHandlers{
	//	actions.POST: ,
	//	actions.PUT: ,
	//})
	c.pathTrie.insert("/{index}/{type}/{id}/_source", actions.MethodHandlers{
		actions.GET: actions.NewRestGetSource(clusterService, indicesService, indexNameExpressionResolver, transportService),
	})

	s := &fasthttp.Server{
		Handler: c.HandleFastHTTP,
	}
	return &Bootstrap{
		s: s,
	}
}

func (b *Bootstrap) Start(port string) error {
	return b.s.ListenAndServe(port)
}
