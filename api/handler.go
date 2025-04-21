package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Miku7676/Raft3D/raft"
	"github.com/Miku7676/Raft3D/store"
	"github.com/go-chi/chi"
	hashicraft "github.com/hashicorp/raft"
)

func RegisterRoutes(r chi.Router, raftNode *hashicraft.Raft, fsm *raft.FSM) {
	r.Post("/join", func(w http.ResponseWriter, r *http.Request) {
		joinHandler(w, r, raftNode)
	})

	r.Get("/leader", func(w http.ResponseWriter, r *http.Request) {
		leaderHandler(w, r, raftNode)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/printers", func(w http.ResponseWriter, r *http.Request) {
			createPrintersHandler(w, r, raftNode)
		})
		r.Get("/printers", func(w http.ResponseWriter, r *http.Request) {
			getPrintersHandler(w, r, fsm)
		})

		r.Post("/filaments", func(w http.ResponseWriter, r *http.Request) {
			createFilamentsHandler(w, r, raftNode)
		})
		r.Get("/filaments", func(w http.ResponseWriter, r *http.Request) {
			getFilamentsHandler(w, r, fsm)
		})

		r.Post("/print_jobs", func(w http.ResponseWriter, r *http.Request) {
			createPrintJobsHandler(w, r, raftNode)
		})
		r.Get("/print_jobs", func(w http.ResponseWriter, r *http.Request) {
			getPrintJobsHandler(w, r, fsm)
		})

		r.Post("/print_jobs/{job_id}/status", func(w http.ResponseWriter, r *http.Request) {
			updatePrintJobStatusHandler(w, r, raftNode)
		})
	})
}

func leaderHandler(w http.ResponseWriter, r *http.Request, raftNode *hashicraft.Raft) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(raftNode.Leader()))
}

func joinHandler(w http.ResponseWriter, r *http.Request, raftNode *hashicraft.Raft) {
	if raftNode.State() != hashicraft.Leader {
        http.Error(w, "Not leader", http.StatusForbidden)
        return
    }
	nodeID := r.URL.Query().Get("nodeID")
	addr := r.URL.Query().Get("addr")
	if err := raftNode.AddVoter(hashicraft.ServerID(nodeID), hashicraft.ServerAddress(addr), 0, 0).Error(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	log.Printf("Node %s joined at %s", nodeID, addr)
}

func createPrintersHandler(w http.ResponseWriter, r *http.Request, raftNode *hashicraft.Raft) {
	if raftNode.State() != hashicraft.Leader {
		http.Error(w, "Not the leader", http.StatusForbidden)
		return
	}
	var req store.Printer
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	// for _, entry := range req.Entries {
	// 	payload, _ := json.Marshal(entry.Value)
	// 	cmd, _ := json.Marshal(store.Command{Type: store.AddPrinter, Payload: payload})
	// 	f := raftNode.Apply(cmd, 5*time.Second)
	// 	if err := f.Error(); err != nil {
	// 		log.Println("Apply failed:", err)
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// }

	// for _, entry := range req.Entries {
	// 	// Make sure the ID field is set from the Key if needed
	// 	printer := entry.Value
	// 	if printer.ID == "" {
	// 		printer.ID = entry.Key
	// 	}
	//
	// 	payload, _ := json.Marshal(printer)
	// 	cmd, _ := json.Marshal(store.Command{Type: store.AddPrinter, Payload: payload})
	//
	// 	f := raftNode.Apply(cmd, 5*time.Second)
	// 	if err := f.Error(); err != nil {
	// 		log.Println("Apply failed:", err)
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// }

	if req.ID == "" {
		http.Error(w, "Printer ID is required", http.StatusBadRequest)
		return
	}

	payload, _ := json.Marshal(req)
	cmd, _ := json.Marshal(store.Command{Type: store.AddPrinter, Payload: payload})
	f := raftNode.Apply(cmd, 5*time.Second)
	if err := f.Error(); err != nil {
		log.Println("Apply failed:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getPrintersHandler(w http.ResponseWriter, r *http.Request, fsm *raft.FSM) {
	fsm.Mu.Lock()
	defer fsm.Mu.Unlock()

	response := store.Printers{
		Entries: make([]struct {
			Key   string        `json:"key"`
			Value store.Printer `json:"value"`
		}, 0, len(fsm.Printers)),
	}

	for id, printer := range fsm.Printers {
		response.Entries = append(response.Entries, struct {
			Key   string        `json:"key"`
			Value store.Printer `json:"value"`
		}{
				Key:   id,
				Value: printer,
			})
	}

	json.NewEncoder(w).Encode(response)
}

func createFilamentsHandler(w http.ResponseWriter, r *http.Request, raftNode *hashicraft.Raft) {
	if raftNode.State() != hashicraft.Leader {
		http.Error(w, "Not the leader", http.StatusForbidden)
		return
	}
	var req store.Filament
	json.NewDecoder(r.Body).Decode(&req)

	// for _, entry := range req.Entries {
	// 	payload, _ := json.Marshal(entry.Value)
	// 	cmd, _ := json.Marshal(store.Command{Type: store.AddFilament, Payload: payload})
	// 	f := raftNode.Apply(cmd, 5*time.Second)
	// 	if err := f.Error(); err != nil {
	// 		log.Println("Apply failed:", err)
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// }

	if req.ID == "" {
		http.Error(w, "Printer ID is required", http.StatusBadRequest)
		return
	}

	payload, _ := json.Marshal(req)
	cmd, _ := json.Marshal(store.Command{Type: store.AddFilament, Payload: payload})
	f := raftNode.Apply(cmd, 5*time.Second)
	if err := f.Error(); err != nil {
		log.Println("Apply failed:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getFilamentsHandler(w http.ResponseWriter, r *http.Request, fsm *raft.FSM) {
	fsm.Mu.Lock()
	defer fsm.Mu.Unlock()
	response := store.Filaments{
		Entries: make([]struct {
			Key   string        `json:"key"`
			Value store.Filament `json:"value"`
		}, 0, len(fsm.Filaments)),
	}

	for id, filament := range fsm.Filaments {
		response.Entries = append(response.Entries, struct {
			Key   string        `json:"key"`
			Value store.Filament `json:"value"`
		}{
				Key:   id,
				Value: filament,
			})
	}

	json.NewEncoder(w).Encode(response)

}

func createPrintJobsHandler(w http.ResponseWriter, r *http.Request, raftNode *hashicraft.Raft) {
	if raftNode.State() != hashicraft.Leader {
		http.Error(w, "Not the leader", http.StatusForbidden)
		return
	}
	var req store.PrintJob

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	// for _, entry := range req.Entries {
	// 	payload, _ := json.Marshal(entry.Value)
	// 	cmd, _ := json.Marshal(store.Command{Type: store.AddJob, Payload: payload})
	// 	f := raftNode.Apply(cmd, 5*time.Second)
	// 	if err := f.Error(); err != nil {
	// 		log.Println("Apply failed:", err)
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// }

	if req.ID == "" {
		http.Error(w, "Printer ID is required", http.StatusBadRequest)
		return
	}

	payload, _ := json.Marshal(req)
	cmd, _ := json.Marshal(store.Command{Type: store.AddJob, Payload: payload})
	f := raftNode.Apply(cmd, 5*time.Second)
	if err := f.Error(); err != nil {
		log.Println("Apply failed:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getPrintJobsHandler(w http.ResponseWriter, r *http.Request, fsm *raft.FSM) {
	fsm.Mu.Lock()
	defer fsm.Mu.Unlock()

	response := store.PrintJobs{
		Entries: make([]struct {
			Key   string        `json:"key"`
			Value store.PrintJob `json:"value"`
		}, 0, len(fsm.Jobs)),
	}

	for id, job := range fsm.Jobs {
		response.Entries = append(response.Entries, struct {
			Key   string        `json:"key"`
			Value store.PrintJob `json:"value"`
		}{
				Key:   id,
				Value: job,
			})
	}

	json.NewEncoder(w).Encode(response)
}

func updatePrintJobStatusHandler(w http.ResponseWriter, r *http.Request, raftNode *hashicraft.Raft) {
	if raftNode.State() != hashicraft.Leader {
		http.Error(w, "Not the leader", http.StatusForbidden)
		return
	}
	jobID := chi.URLParam(r, "job_id")
	newStatus := r.URL.Query().Get("status")
	if jobID == "" || newStatus == "" {
		http.Error(w, "Missing job_id or status", http.StatusBadRequest)
		return
	}

	update := store.PrintJob{
		ID:     jobID,
		Status: newStatus,
	}

	payload, _ := json.Marshal(update)
	cmd, _ := json.Marshal(store.Command{Type: store.UpdateJob, Payload: payload})
	f := raftNode.Apply(cmd, 5*time.Second)
	if err := f.Error(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

