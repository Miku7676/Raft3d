package raft

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	hraft "github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
)

// func SetupRaft(nodeID, bindAddr, dataDir string, fsm *FSM, join bool, joinAddr string) (*hashicraft.Raft,  error) { //hashicraft.Transport,
// 	config := hashicraft.DefaultConfig()
// 	config.LocalID = hashicraft.ServerID(nodeID)
//
// 	os.MkdirAll(dataDir, 0700)
// 	addr, _ := net.ResolveTCPAddr("tcp", bindAddr)
// 	transport, _ := hashicraft.NewTCPTransport(bindAddr, addr, 3, 10*time.Second, os.Stdout)
//
// 	snapshots, _ := hashicraft.NewFileSnapshotStore(dataDir, 1, os.Stdout)
// 	logStore, _ := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-log.bolt"))
// 	stableStore, _ := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-stable.bolt"))
//
// 	r, err := hashicraft.NewRaft(config, fsm, logStore, stableStore, snapshots, transport)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if !join {
// 		r.BootstrapCluster(hashicraft.Configuration{
// 			Servers: []hashicraft.Server{{ID: config.LocalID, Address: transport.LocalAddr()}},
// 		})
// 	} else {
// 		// Join another cluster
// 		_, err := http.Post(fmt.Sprintf("http://%s/join?nodeID=%s&addr=%s", joinAddr, nodeID, bindAddr), "", nil)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
//
// 	return r, nil
// }

func SetupRaft(nodeID, bindAddr, dataDir string, fsm *FSM, join bool, joinAddr string) (*hraft.Raft, *hraft.NetworkTransport, error) {
	config := hraft.DefaultConfig()
	config.LocalID = hraft.ServerID(nodeID)

	os.MkdirAll(dataDir, 0700)

	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1"+bindAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve TCP addr: %w", err)
	}

	transport, err := hraft.NewTCPTransport(bindAddr, addr, 3, 10*time.Second, os.Stdout)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create transport: %w", err)
	}

	snapshots, err := hraft.NewFileSnapshotStore(dataDir, 1, os.Stdout)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create snapshot store: %w", err)
	}

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-log.bolt"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create log store: %w", err)
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-stable.bolt"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stable store: %w", err)
	}

	r, err := hraft.NewRaft(config, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create raft node: %w", err)
	}

	if !join {
		r.BootstrapCluster(hraft.Configuration{
			Servers: []hraft.Server{{ID: config.LocalID, Address: transport.LocalAddr()}},
		})
	} else {
		// Join to existing cluster
		_, err := http.Post(fmt.Sprintf("http://%s/join?nodeID=%s&addr=%s", joinAddr, nodeID, bindAddr), "", nil)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to join cluster: %w", err)
		}
	}

	return r, transport, nil
}

