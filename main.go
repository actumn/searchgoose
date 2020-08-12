package main

import (
	"github.com/actumn/searchgoose/http"
	"github.com/actumn/searchgoose/state"
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
	id := cluster.GenerateNodeId()
	transportService := transport.Service{
		LocalNode: state.CreateLocalNode(id),
	}
	clusterService := cluster.Service{}
	persistClusterStateService := persist.ClusterStateService{}

	gateway := metadata.GatewayMetaState{}
	gateway.Start(
		transportService,
		clusterService,
		persistClusterStateService,
	)

	coordinator := discovery.Coordinator{
		TransportService:      transportService,
		ClusterApplierService: clusterService.ApplierService,
	}
	coordinator.Start()

	b := http.New(&clusterService)
	log.Println("start server...")
	if err := b.Start(); err != nil {
		panic(err)
	}
}
