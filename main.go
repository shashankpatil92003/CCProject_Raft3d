package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"raft3d/api"
	raftnode "raft3d/raft"

	"github.com/hashicorp/raft"
)

func main() {
	nodeID := flag.String("id", "node1", "Unique ID for this node")
	httpAddr := flag.String("http", "8080", "HTTP port to listen on")
	raftDir := flag.String("raft", "./data/node1", "Raft storage directory")
	raftPort := flag.String("raftport", "5001", "Port for Raft TCP communication")
	bootstrap := flag.Bool("bootstrap", false, "Set to true to bootstrap the cluster")
	flag.Parse()

	os.Setenv("RAFT_PORT", *raftPort)

	fmt.Printf("Starting node %s on HTTP :%s, Raft Dir: %s, Bootstrap: %v\n",
		*nodeID, *httpAddr, *raftDir, *bootstrap)

	r, fsm, addr, id := raftnode.SetupRaftNode(*nodeID, *raftDir, *bootstrap)
	fmt.Printf("Raft transport address: %s\n", addr)

	api.RegisterRaft(r)
	api.RegisterFSM(fsm)

	go api.StartServer(*httpAddr)

	// Periodically log current state
	go func() {
		for {
			time.Sleep(5 * time.Second)
			state := r.State()
			leader := r.Leader()
			if state == raft.Leader {
				fmt.Printf("[INFO] I (%s) am the Leader\n", id)
			} else {
				fmt.Printf("[INFO] I (%s) am a Follower. Current leader: %s\n", id, leader)
			}
		}
	}()

	select {} // keep alive
}
