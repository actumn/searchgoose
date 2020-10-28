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
	NodesInfoAction  = "cluster:monitor/nodes/info"
	NodesStatsAction = "cluster:monitor/nodes/stats"
)

type RestNodesInfo struct {
	clusterService   *cluster.Service
	transportService *transport.Service
}

func NewRestNodesInfo(clusterService *cluster.Service, transportService *transport.Service) *RestNodesInfo {
	//monitorService := monitor.NewService()
	//transportService.RegisterRequestHandler(NodesInfoAction, func(channel transport.ReplyChannel, req []byte) {
	//
	//})

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
	transportService.RegisterRequestHandler(NodesStatsAction, func(channel transport.ReplyChannel, req []byte) {
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
		h.transportService.SendRequest(node, NodesStatsAction, []byte(""), func(response []byte) {
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
			"jvm": map[string]interface{}{
				"mem": map[string]interface{}{
					"heap_used_in_bytes": response.NodeStats.RuntimeStats.HeapAlloc,
					"heap_max_in_bytes":  response.NodeStats.RuntimeStats.HeapSys,
					"heap_used_percent":  response.NodeStats.RuntimeStats.HeapAlloc * 100 / response.NodeStats.RuntimeStats.HeapSys,
				},
			},
			"os": map[string]interface{}{
				"cpu": map[string]interface{}{
					"percent": response.NodeStats.OsStats.CpuStats.Percent,
					"load_average": map[string]interface{}{
						"1m":  response.NodeStats.OsStats.CpuStats.LoadAverage.Load1,
						"5m":  response.NodeStats.OsStats.CpuStats.LoadAverage.Load5,
						"15m": response.NodeStats.OsStats.CpuStats.LoadAverage.Load15,
					},
				},
				"mem": map[string]interface{}{
					"total_in_bytes": response.NodeStats.OsStats.MemStats.Total,
					"free_in_bytes":  response.NodeStats.OsStats.MemStats.Free,
					"used_in_bytes":  response.NodeStats.OsStats.MemStats.Total - response.NodeStats.OsStats.MemStats.Free,
					"free_percent":   response.NodeStats.OsStats.MemStats.Free * 100 / response.NodeStats.OsStats.MemStats.Total,
					"used_percent":   100 - response.NodeStats.OsStats.MemStats.Free*100/response.NodeStats.OsStats.MemStats.Total,
				},
			},
			"fs": map[string]interface{}{
				"total": map[string]interface{}{
					"total_in_bytes":     response.NodeStats.FsStats.Total,
					"free_in_bytes":      response.NodeStats.FsStats.Free,
					"available_in_bytes": response.NodeStats.FsStats.Available,
				},
			},
			"process": map[string]interface{}{
				"open_file_descriptors": response.NodeStats.ProcStats.NumFDs,
				"max_file_descriptors":  -1,
				"cpu": map[string]interface{}{
					"percent": response.NodeStats.ProcStats.CpuPercent,
				},
				"mem": map[string]interface{}{
					"total_virtual_in_bytes": response.NodeStats.ProcStats.MemTotalVirtual,
				},
			},
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
