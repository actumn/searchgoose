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

	viper.SetConfigName("searchgoose") // config file name without extension
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Warn("error finding config file: searchgoose", err)
	}

	seedHosts := flag.String("seed_hosts", "", "연결할 노드들")
	host := flag.String("host_address", "0.0.0.0", "호스트 주소")
	tcpPort := flag.Int("transport.port", 8180, "Transport 연결 노드")
	httpPort := flag.Int("http.port", 8080, "HTTP 연결 노드")
	nodeName := flag.String("node.name", "", "노드 이름")

	flag.Parse()

	viper.Set("discovery.seed_hosts", *seedHosts)
	viper.Set("network.host", *host)
	viper.Set("transport.port", *tcpPort)
	viper.Set("http.port", *httpPort)
	viper.Set("node.name", *nodeName)

	nodeId := cluster.GenerateNodeId()
	logrus.Info("[Node Id]: ", nodeId)
	viper.Set("node.id", nodeId)

}

func start() {

	// for signal handling
	var outer chan int
	var count, length int
	outer = make(chan int, 1)
	done := func() {
		outer <- 1
	}

	var tcpTransport transport.Transport

	tcpPort := viper.GetInt("transport.port")
	seedHost := viper.GetString("discovery.seed_hosts")
	id := viper.GetString("node.id")
	name := viper.GetString("node.name")

	tcpTransport = tcp.NewTransport(tcpPort, seedHost, id)
	length = len(tcpTransport.GetSeedHosts())
	transportService := transport.NewService(tcpTransport, name)
	transportService.Start(tcpPort)

	clusterService := cluster.NewService()
	persistClusterStateService := persist.NewClusterStateService()

	gateway := metadata.NewGatewayMetaState()
	gateway.Start(transportService, clusterService, persistClusterStateService)

	coordinator := discovery.NewCoordinator(transportService, clusterService.ApplierService, clusterService.MasterService, gateway.PersistedState)
	coordinator.Done = done

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
	coordinator.StartInitialJoin()

	for i := 0; i < length; i++ {
		<-outer
		count++
	}

	indexNameExpressionResolver := indices.NewNameExpressionResolver()

	b := http.New(clusterService, clusterMetadataCreateIndexService, clusterMetadataDeleteIndexService, clusterMetadataIndexAliasService, indicesService, transportService, indexNameExpressionResolver)
	httpPort := ":" + viper.GetString("http.port")

	if count == length {
		logrus.Printf("124 leader=%v")
		logrus.Info("start server...")
		coordinator.Started = true
		if err := b.Start(httpPort); err != nil {
			panic(err)
		}
	}
}
