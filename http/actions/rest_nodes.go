package actions

import (
	"bytes"
	"encoding/gob"
	"github.com/actumn/searchgoose/monitor"
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

const (
	NodesInfoAction  = "cluster:monitor/nodes/info"
	NodesStatsAction = "cluster:monitor/nodes/stats"
)

type nodeInfoResponse struct {
	Node     state.Node
	NodeInfo monitor.Info
}

func (r *nodeInfoResponse) toBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func nodeInfoResponseFromBytes(b []byte) *nodeInfoResponse {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var req nodeInfoResponse
	if err := decoder.Decode(&req); err != nil {
		logrus.Fatal(err)
	}
	return &req
}

type RestNodesInfo struct {
	clusterService   *cluster.Service
	transportService *transport.Service
}

func NewRestNodesInfo(clusterService *cluster.Service, transportService *transport.Service) *RestNodesInfo {
	monitorService := monitor.NewService()
	transportService.RegisterRequestHandler(NodesInfoAction, func(channel transport.ReplyChannel, req []byte) {
		info := monitorService.Info()

		nodeRes := nodeInfoResponse{
			Node:     *transportService.LocalNode,
			NodeInfo: info,
		}
		channel.SendMessage("", nodeRes.toBytes())
	})

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

	responses := make([]nodeInfoResponse, len(nodes))
	wg := sync.WaitGroup{}
	wg.Add(len(nodes))
	idx := -1
	for _, node := range nodes {
		idx += 1
		currIdx := idx
		h.transportService.SendRequest(node, NodesInfoAction, []byte(""), func(response []byte) {
			nodeStatsRes := nodeInfoResponseFromBytes(response)
			responses[currIdx] = *nodeStatsRes
			wg.Done()
		})
	}
	wg.Wait()

	wd, _ := os.Getwd()
	nodesMap := map[string]interface{}{}
	for _, response := range responses {
		nodesMap[response.Node.Id] = map[string]interface{}{
			"transport_address": response.Node.HostAddress,
			"host":              response.Node.HostAddress,
			"ip":                response.Node.HostAddress,
			"version":           "7.8.1",
			"roles":             []string{"master", "data"},
			"settings": map[string]interface{}{
				"node": map[string]interface{}{
					"master": true,
					"data":   true,
				},
				"path": map[string]interface{}{
					"data": []string{
						wd + "/data",
					},
					"logs": wd + "/logs",
					"home": wd,
				},
			},
			"os": map[string]interface{}{
				"name":                 response.NodeInfo.Os.Name,
				"pretty_name":          response.NodeInfo.Os.Platform,
				"arch":                 response.NodeInfo.Os.Arch,
				"version":              response.NodeInfo.Os.Version,
				"available_processors": response.NodeInfo.Os.Processors,
				"allocated_processors": response.NodeInfo.Os.Processors,
			},
			"jvm": map[string]interface{}{
				"pid":     response.NodeInfo.Runtime.Pid,
				"version": response.NodeInfo.Runtime.Version,
			},
			"http": map[string]interface{}{
				"publish_address": response.Node.HostAddress,
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
			"nodes":        nodesMap,
		},
	})
}

type nodeStatsResponse struct {
	Node         state.Node
	NodeStats    monitor.Stats
	IndicesStats indices.Stats
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

func NewRestNodesStats(clusterService *cluster.Service, indicesService *indices.Service, transportService *transport.Service) *RestNodesStats {
	monitorService := monitor.NewService()
	transportService.RegisterRequestHandler(NodesStatsAction, func(channel transport.ReplyChannel, req []byte) {
		stats := monitorService.Stats()
		// TODO :: find out more elastic way to compute total indices stats
		indicesStats := indices.Stats{
			NumDocs:    0,
			NumDeleted: 0,
			NumBytes:   0,
		}

		for _, indexService := range indicesService.Indices {
			for _, shard := range indexService.Shards {
				shardStats := shard.Stats()

				indicesStats.NumDocs += shardStats.NumDocs
				indicesStats.NumDeleted += shardStats.UserData["deletes"].(uint64)
				indicesStats.NumBytes += shardStats.UserData["num_bytes_used_disk"].(uint64)
			}
		}

		nodeRes := nodeStatsResponse{
			Node:         *transportService.LocalNode,
			NodeStats:    stats,
			IndicesStats: indicesStats,
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

	wd, _ := os.Getwd()
	nodeStatsMap := map[string]interface{}{}
	for _, response := range responses {
		nodeStatsMap[response.Node.Id] = map[string]interface{}{
			"name":              response.Node.Name + response.Node.Id,
			"transport_address": response.Node.HostAddress,
			"host":              response.Node.HostAddress,
			"ip":                response.Node.HostAddress,
			"roles":             []string{"data", "master"},
			"indices": map[string]interface{}{
				"docs": map[string]interface{}{
					"count":   response.IndicesStats.NumDocs,
					"deleted": response.IndicesStats.NumDeleted,
				},
				"store": map[string]interface{}{
					"size_in_bytes": response.IndicesStats.NumBytes,
				},
			},
			"jvm": map[string]interface{}{
				"mem": map[string]interface{}{
					"heap_used_in_bytes": response.NodeStats.Runtime.HeapAlloc,
					"heap_max_in_bytes":  response.NodeStats.Runtime.HeapSys,
					"heap_used_percent":  response.NodeStats.Runtime.HeapAlloc * 100 / response.NodeStats.Runtime.HeapSys,
				},
			},
			"os": map[string]interface{}{
				"cpu": map[string]interface{}{
					"percent": response.NodeStats.Os.Cpu.Percent,
					"load_average": map[string]interface{}{
						"1m":  response.NodeStats.Os.Cpu.LoadAverage.Load1,
						"5m":  response.NodeStats.Os.Cpu.LoadAverage.Load5,
						"15m": response.NodeStats.Os.Cpu.LoadAverage.Load15,
					},
				},
				"mem": map[string]interface{}{
					"total_in_bytes": response.NodeStats.Os.Mem.Total,
					"free_in_bytes":  response.NodeStats.Os.Mem.Free,
					"used_in_bytes":  response.NodeStats.Os.Mem.Total - response.NodeStats.Os.Mem.Free,
					"free_percent":   response.NodeStats.Os.Mem.Free * 100 / response.NodeStats.Os.Mem.Total,
					"used_percent":   100 - response.NodeStats.Os.Mem.Free*100/response.NodeStats.Os.Mem.Total,
				},
			},
			"fs": map[string]interface{}{
				"total": map[string]interface{}{
					"total_in_bytes":     response.NodeStats.Fs.Total,
					"free_in_bytes":      response.NodeStats.Fs.Free,
					"available_in_bytes": response.NodeStats.Fs.Available,
				},
				"data": []map[string]interface{}{
					{
						"path":               wd + "/data",
						"total_in_bytes":     response.NodeStats.Fs.Total,
						"free_in_bytes":      response.NodeStats.Fs.Free,
						"available_in_bytes": response.NodeStats.Fs.Available,
					},
				},
			},
			"process": map[string]interface{}{
				"open_file_descriptors": response.NodeStats.Proc.NumFDs,
				"max_file_descriptors":  -1,
				"cpu": map[string]interface{}{
					"percent": response.NodeStats.Proc.CpuPercent,
				},
				"mem": map[string]interface{}{
					"total_virtual_in_bytes": response.NodeStats.Proc.MemTotalVirtual,
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
