package raftnode

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

func SetupRaftNode(nodeID, raftDir string, bootstrap bool) (*raft.Raft, *PrinterFSM, raft.ServerAddress, raft.ServerID) {
	rand.Seed(time.Now().UnixNano())

	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)

	// Timeouts
	config.HeartbeatTimeout = 100 * time.Millisecond
	config.ElectionTimeout = time.Duration(150+rand.Intn(150)) * time.Millisecond
	config.LeaderLeaseTimeout = 100 * time.Millisecond
	config.CommitTimeout = 50 * time.Millisecond

	// Directories
	if err := os.MkdirAll(raftDir, 0700); err != nil {
		log.Fatalf("failed to create raft directory: %v", err)
	}

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDir, "raft-log.bolt"))
	if err != nil {
		log.Fatalf("failed to create log store: %v", err)
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDir, "raft-stable.bolt"))
	if err != nil {
		log.Fatalf("failed to create stable store: %v", err)
	}

	snapshotStore, err := raft.NewFileSnapshotStore(raftDir, 1, os.Stdout)
	if err != nil {
		log.Fatalf("failed to create snapshot store: %v", err)
	}

	addr := "127.0.0.1:" + os.Getenv("RAFT_PORT")
	transport, err := raft.NewTCPTransport(addr, nil, 3, 10*time.Second, os.Stdout)
	if err != nil {
		log.Fatalf("failed to create TCP transport: %v", err)
	}

	log.Printf("Transport started at address: %s\n", addr)

	fsm := NewPrinterFSM()

	r, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		log.Fatalf("failed to create raft: %v", err)
	}

	if bootstrap {
		initialConfig := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		if err := r.BootstrapCluster(initialConfig).Error(); err != nil {
			log.Fatalf("failed to bootstrap cluster: %v", err)
		}
		log.Println("Raft node initialized as bootstrap node with address:", transport.LocalAddr())
	}

	return r, fsm, raft.ServerAddress(fmt.Sprintf("%v", transport.LocalAddr())), config.LocalID
}
