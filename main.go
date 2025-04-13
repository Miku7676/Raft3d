package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/Miku7676/Raft3D/api"
	"github.com/Miku7676/Raft3D/raft"
)

func main() {
	nodeID := flag.String("id", "node1", "Unique node ID")
	httpAddr := flag.String("http", ":8080", "HTTP server address")
	raftAddr := flag.String("raft", ":9001", "Raft communication address")
	dataDir := flag.String("data", "/tmp/raft3d", "Raft data directory")
	joinAddr := flag.String("join", "", "Join address if not bootstrapping")
	flag.Parse()

	fsm := raft.NewFSM()
	join := *joinAddr != ""

	raftNode, _, err := raft.SetupRaft(*nodeID, *raftAddr, *dataDir, fsm, join, *joinAddr) 
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	api.RegisterRoutes(r, raftNode, fsm)

	cfgFuture := raftNode.GetConfiguration()
	if err := cfgFuture.Error(); err == nil {
		for _, server := range cfgFuture.Configuration().Servers {
			fmt.Println("Cluster member:", server.ID, server.Address)
		}
	} 
	fmt.Printf("[%s] HTTP server listening at %s\n", *nodeID, *httpAddr)
	http.ListenAndServe(*httpAddr, r)
}
