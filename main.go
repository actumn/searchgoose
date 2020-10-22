package main

import (
	"flag"
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
	"github.com/spf13/viper"
	"runtime"
	"strings"
	"time"
)

func main() {
	start()
}

func init() {
	logrus.SetLevel(logrus.TraceLevel)
	//logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			functionName := frame.Function[strings.LastIndex(frame.Function, ".")+1:]
			fileName := frame.File[strings.LastIndex(frame.File, "/")+1:]
			return fmt.Sprintf("%-20s", functionName+"()"), fmt.Sprintf("%s:%d\t", fileName, frame.Line)
		},
	})

	viper.SetConfigName("searchgoose") // config file name without extension
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatal("fatal error config file: searchgoose", err)
	}

	seedHosts := flag.String("seed_hosts", "", "연결할 노드들")
	host := flag.String("host_address", "", "호스트 주소")
	tcpPort := flag.Int("transport.port", 0, "Transport 연결 노드")
	httpPort := flag.Int("http.port", 0, "HTTP 연결 노드")

	flag.Parse()

	viper.Set("discovery.seed_hosts", *seedHosts)
	viper.Set("network.host", *host)
	viper.Set("transport.port", *tcpPort)
	viper.Set("http.port", *httpPort)

	nodeId := cluster.GenerateNodeId()
	logrus.Info("[Node Id]: ", nodeId)
	viper.Set("node.id", nodeId)

}

func start() {

	var tcpTransport transport.Transport

	tcpTransport = tcp.NewTransport()
	transportService := transport.NewService(tcpTransport)
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
	clusterMetadataDeleteIndexService := cluster.NewMetadataDeleteIndexService(clusterService, allocationService)
	clusterMetadataIndexAliasService := cluster.NewMetadataIndexAliasService(clusterService)

	gateway.Start(transportService, clusterService, persistClusterStateService)

	coordinator.Start()
	time.Sleep(time.Duration(15) * time.Second)

	coordinator.StartInitialJoin()
	time.Sleep(time.Duration(1000) * time.Second)
	indexNameExpressionResolver := indices.NewNameExpressionResolver()

	b := http.New(clusterService, clusterMetadataCreateIndexService, clusterMetadataDeleteIndexService, clusterMetadataIndexAliasService, indicesService, transportService, indexNameExpressionResolver)
	httpPort := ":" + viper.GetString("http.port")
	logrus.Info("start server...")
	if err := b.Start(httpPort); err != nil {
		panic(err)
	}
}
