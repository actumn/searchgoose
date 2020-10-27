package actions

import (
	"bytes"
	"encoding/gob"
	"github.com/actumn/searchgoose/monitor"
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/sirupsen/logrus"
	"sync"
)

const (
	NodeStatsAction = "cluster:monitor/nodes/stats"
)

type RestNodesInfo struct {
	clusterService   *cluster.Service
	transportService *transport.Service
}

func NewRestNodesInfo(clusterService *cluster.Service, transportService *transport.Service) *RestNodesInfo {
	return &RestNodesInfo{
		clusterService:   clusterService,
		transportService: transportService,
	}
}

func (h *RestNodesInfo) Handle(r *RestRequest, reply ResponseListener) {
	clusterState := h.clusterService.State()
	nodes := clusterState.Nodes.Nodes
	nodeId := r.PathParams["nodeId"]
	var concreteNodes []state.Node

	if nodeId == "" {
		for _, node := range nodes {
			concreteNodes = append(concreteNodes, node)
		}
	} else if node, existing := nodes[nodeId]; existing {
		concreteNodes = []state.Node{node}
	} else {
		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"_nodes": map[string]interface{}{
					"total":      0,
					"successful": 0,
					"failed":     0,
				},
				"cluster_name": clusterState.Name,
				"nodes":        map[string]interface{}{},
			},
		})
		return
	}

	nodesMap := map[string]interface{}{}
	for _, node := range concreteNodes {
		nodesMap[node.Id] = map[string]interface{}{
			"ip":      node.HostAddress,
			"version": "7.8.1",
			"http": map[string]interface{}{
				"public_address": node.HostAddress,
			},
		}
	}

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"nodes": nodesMap,
		},
	})
}

type nodeStatsResponse struct {
	Node      state.Node
	NodeStats monitor.Stats
}

func (r *nodeStatsResponse) toBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func nodeStatsResponseFromBytes(b []byte) *nodeStatsResponse {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var req nodeStatsResponse
	if err := decoder.Decode(&req); err != nil {
		logrus.Fatal(err)
	}
	return &req
}

type RestNodesStats struct {
	clusterService   *cluster.Service
	transportService *transport.Service
}

func NewRestNodesStats(clusterService *cluster.Service, transportService *transport.Service) *RestNodesStats {
	monitorService := monitor.NewService()
	transportService.RegisterRequestHandler(NodeStatsAction, func(channel transport.ReplyChannel, req []byte) {
		stats := monitorService.Stats()

		nodeRes := nodeStatsResponse{
			Node:      *transportService.LocalNode,
			NodeStats: stats,
		}
		channel.SendMessage("", nodeRes.toBytes())
	})

	return &RestNodesStats{
		clusterService:   clusterService,
		transportService: transportService,
	}
}

func (h *RestNodesStats) Handle(r *RestRequest, reply ResponseListener) {
	clusterState := h.clusterService.State()
	nodes := clusterState.Nodes.Nodes
	nodeId := r.PathParams["nodeId"]
	var concreteNodes []state.Node

	if nodeId == "" {
		for _, node := range nodes {
			concreteNodes = append(concreteNodes, node)
		}
	} else if node, existing := nodes[nodeId]; existing {
		concreteNodes = []state.Node{node}
	} else {
		reply(RestResponse{
			StatusCode: 200,
			Body: map[string]interface{}{
				"_nodes": map[string]interface{}{
					"total":      0,
					"successful": 0,
					"failed":     0,
				},
				"cluster_name": clusterState.Name,
				"nodes":        map[string]interface{}{},
			},
		})
		return
	}

	responses := make([]nodeStatsResponse, len(nodes))
	wg := sync.WaitGroup{}
	wg.Add(len(nodes))
	idx := -1
	for _, node := range nodes {
		idx += 1
		currIdx := idx
		h.transportService.SendRequest(node, NodeStatsAction, []byte(""), func(response []byte) {
			nodeStatsRes := nodeStatsResponseFromBytes(response)
			responses[currIdx] = *nodeStatsRes
			wg.Done()
		})
	}
	wg.Wait()

	nodeStatsMap := map[string]interface{}{}
	for _, response := range responses {
		nodeStatsMap[response.Node.Id] = map[string]interface{}{
			"transport_address": response.Node.HostAddress,
			"host":              response.Node.HostAddress,
			"ip":                response.Node.HostAddress,
			"roles":             []string{"data", "master"},
			"os":                map[string]interface{}{},
		}
	}

	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"_nodes": map[string]interface{}{
				"total":      len(nodes),
				"successful": len(nodes),
				"failed":     0,
			},
			"cluster_name": clusterState.Name,
			"nodes":        nodeStatsMap,
		},
	})
}
