package main

import (
	"fmt"
	"github.com/actumn/searchgoose/http"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/discovery"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/actumn/searchgoose/state/metadata"
	"github.com/actumn/searchgoose/state/persist"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/actumn/searchgoose/state/transport/tcp"
	"github.com/sirupsen/logrus"
	"runtime"
	"strings"
)

func main() {
	start()
}

func init() {
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			functionName := frame.Function[strings.LastIndex(frame.Function, ".")+1:]
			fileName := frame.File[strings.LastIndex(frame.File, "/")+1:]
			return fmt.Sprintf("%-20s", functionName+"()"), fmt.Sprintf("%s:%d\t", fileName, frame.Line)
		},
	})
}

func start() {
	nodeId := cluster.GenerateNodeId()
	logrus.Info("[Node Id]: ", nodeId)

	// TODO :: 현재 노드의 host와 port 설정을 가져오게 하자
	address := "localhost:8180"
	//address := "localhost:8179"
	//address := "localhost:8181"

	seedHosts := []string{"localhost:8179", "localhost:8181"} //8180
	//seedHosts := []string{"localhost:8180"} //8179
	//seedHosts := []string{"localhost:8180"} //8181

	var tcpTransport transport.Transport

	tcpTransport = tcp.NewTransport(address, nodeId, seedHosts)
	transportService := transport.NewService(nodeId, tcpTransport)
	transportService.Start()

	clusterService := cluster.NewService()
	persistClusterStateService := persist.NewClusterStateService()

	gateway := metadata.NewGatewayMetaState()
	gateway.Start(transportService, clusterService, persistClusterStateService)

	coordinator := discovery.NewCoordinator(transportService, clusterService.ApplierService, clusterService.MasterService, gateway.PersistedState)

	indicesService := indices.NewService()
	indicesClusterStateService := indices.NewClusterStateService(indicesService)

	clusterService.ApplierService.AddApplier(indicesClusterStateService.ApplyClusterState)
	clusterService.MasterService.ClusterStatePublish = coordinator.Publish

	allocationService := cluster.NewAllocationService()
	clusterMetadataCreateIndexService := cluster.NewMetadataCreateIndexService(clusterService, allocationService)

	gateway.Start(transportService, clusterService, persistClusterStateService)

	coordinator.Start()
	//time.Sleep(time.Duration(15) * time.Second)
	//
	//coordinator.StartInitialJoin()
	//time.Sleep(time.Duration(1000) * time.Second)
	indexNameExpressionResolver := indices.NewNameExpressionResolver()

	b := http.New(clusterService, clusterMetadataCreateIndexService, indicesService, transportService, indexNameExpressionResolver)
	logrus.Info("start server...")
	if err := b.Start(":8080"); err != nil {
		panic(err)
	}
	//if err := b.Start(":8081"); err != nil { panic(err) }
	//if err := b.Start(":8082"); err != nil { panic(err) }
}
