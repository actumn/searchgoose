package main

import (
	"github.com/actumn/searchgoose/http"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/discovery"
	"github.com/actumn/searchgoose/state/metadata"
	"github.com/actumn/searchgoose/state/persist"
	"github.com/actumn/searchgoose/state/transport"
	"log"
)

func main() {
	start()
}

func start() {
	nodeId := cluster.GenerateNodeId()

	transportService := transport.NewService(nodeId)
	clusterService := cluster.NewService()
	persistClusterStateService := persist.NewClusterStateService()

	gateway := metadata.NewGatewayMetaState()

	coordinator := discovery.NewCoordinator(transportService, clusterService.ApplierService, clusterService.MasterService)

	clusterService.MasterService.ClusterStatePublish = coordinator.Publish

	gateway.Start(
		transportService,
		clusterService,
		persistClusterStateService,
	)
	coordinator.Start()

	b := http.New(clusterService)
	log.Println("start server...")
	if err := b.Start(); err != nil {
		panic(err)
	}
}
