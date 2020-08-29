package main

import (
	"github.com/actumn/searchgoose/http"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/discovery"
	"github.com/actumn/searchgoose/state/indices"
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
}
