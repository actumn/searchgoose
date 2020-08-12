package main

import (
	"github.com/actumn/searchgoose/http"
	"log"
)

func main() {
	b := http.New()
	log.Println("start server...")
	if err := b.Start(); err != nil {
		panic(err)
	}
}

//func start() {
//	id := cluster.GenerateNodeId()
//	transportService := transport.Service{
//		LocalNode: discovery.CreateLocal(id),
//	}
//	clusterService := cluster.Service{}
//	persistClusterStateService := persist.ClusterStateService{}
//
//	gateway := metadata.GatewayMetaState{}
//	gateway.Start(
//		transportService,
//		clusterService,
//		persistClusterStateService,
//	)
//}
