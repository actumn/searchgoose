package main

import (
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/transport/tcp"
	"log"
	"net"
	"time"
)

func main() {
	start()
}

func start() {

	nodeId := cluster.GenerateNodeId()
	transport := tcp.NewTransport(nodeId)

	// TODO :: Handshake 로직은 어디에서 관리되어야 할까

	log.Printf("[Node Id] : %s\n", nodeId)

	// TODO :: 현재 노드의 host와 port 설정을 가져오게 하자
	//port := "8179"
	//port := "8180"
	port := "8181"

	transport.Start(port)
	time.Sleep(time.Duration(15) * time.Second)

	//seedHosts := []string{"localhost:8180", "localhost:8181"} //8179
	//seedHosts := []string{"localhost:8179", "localhost:8181"} //8180
	seedHosts := []string{"localhost:8179", "localhost:8180"} //8181

	log.Printf("Start handshaking\n")

	connections := make(chan *net.Conn)
	for _, seedHost := range seedHosts {
		// Open connection
		transport.OpenConnection(seedHost, connections)
	}

	// Handshake 동작 보기 위해서 잠시 주석 처리
	/*
		transportService := transport.NewService(nodeId)
		clusterService := cluster.NewService()
		persistClusterStateService := persist.NewClusterStateService()

		gateway := metadata.NewGatewayMetaState()
		gateway.Start(transportService, clusterService, persistClusterStateService)

		coordinator := discovery.NewCoordinator(transportService, clusterService.ApplierService, clusterService.MasterService, gateway.PersistedState)

		indicesService := indices.NewService()
		indicesClusterStateService := indices.NewClusterStateService(indicesService)

		clusterService.ApplierService.AddApplier(indicesClusterStateService.ApplyClusterState)
		clusterService.MasterService.ClusterStatePublish = coordinator.Publish

		clusterMetadataCreateIndexService := cluster.NewMetadataCreateIndexService(clusterService)

		gateway.Start(
			transportService,
			clusterService,
			persistClusterStateService,
		)
		coordinator.Start()

		b := http.New(clusterService, clusterMetadataCreateIndexService)
		log.Println("start server...")
		if err := b.Start(); err != nil {
			panic(err)
		}
	*/
}
