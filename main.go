package main

import (
	"github.com/actumn/searchgoose/http"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/discovery"
	"github.com/actumn/searchgoose/state/indices"
	"github.com/actumn/searchgoose/state/metadata"
	"github.com/actumn/searchgoose/state/persist"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/actumn/searchgoose/state/transport/tcp"
	"log"
)

func main() {
	start()
}

func start() {
	nodeId := cluster.GenerateNodeId()
	log.Printf("[Node Id] : %s\n", nodeId)

	// TODO :: 현재 노드의 host와 port 설정을 가져오게 하자
	address := "localhost:8179"
	//address := "localhost:8180"
	//address := "localhost:8181"

	seedHosts := []string{"localhost:8180"} //8179
	//seedHosts := []string{"localhost:8179", "localhost:8181"} //8180
	//seedHosts := []string{"localhost:8180"} //8181

	var tcpTransport transport.Transport

	tcpTransport = tcp.NewTransport(address, nodeId, seedHosts)
	transportService := transport.NewService(nodeId, tcpTransport)
	//transportService.Start()

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

	//coordinator.Start()
	//coordinator.StartInitialJoin()

	b := http.New(clusterService, clusterMetadataCreateIndexService, indicesService)
	log.Println("start server...")
	if err := b.Start(); err != nil {
		panic(err)
	}
}
