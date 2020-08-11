package main

import (
	"github.com/actumn/searchgoose/http"
	"github.com/actumn/searchgoose/services/cluster"
	"github.com/actumn/searchgoose/services/discovery"
	"github.com/actumn/searchgoose/services/metadata"
	"github.com/actumn/searchgoose/services/persist"
	"github.com/actumn/searchgoose/services/transport"
	"log"
)

func main() {
	b := http.New()
	log.Println("start server...")
	if err := b.Start(); err != nil {
		panic(err)
	}
}

func start() {
	id := cluster.GenerateNodeId()
	transportService := transport.Service{
		LocalNode: discovery.CreateLocal(id),
	}
	clusterService := cluster.Service{}
	persistClusterStateService := persist.ClusterStateService{}

	gateway := metadata.GatewayMetaState{}
	gateway.Start(
		transportService,
		clusterService,
		persistClusterStateService,
	)
}
