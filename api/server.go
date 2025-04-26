package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	raftnode "raft3d/raft"

	"github.com/gorilla/mux"
	"github.com/hashicorp/raft"
)

var raftNode *raft.Raft
var fsm *raftnode.PrinterFSM

func RegisterRaft(r *raft.Raft) {
	raftNode = r
}

func RegisterFSM(f *raftnode.PrinterFSM) {
	fsm = f
}

func StartServer(port string) {
	router := mux.NewRouter()

	router.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Raft3D is alive!")
	}).Methods("GET")

	// CRUD APIs
	router.HandleFunc("/api/v1/printers", createPrinterHandler).Methods("POST")
	router.HandleFunc("/api/v1/printers", getPrintersHandler).Methods("GET")

	router.HandleFunc("/api/v1/filaments", createFilamentHandler).Methods("POST")
	router.HandleFunc("/api/v1/filaments", getFilamentsHandler).Methods("GET")

	router.HandleFunc("/api/v1/print_jobs", createPrintJobHandler).Methods("POST")
	router.HandleFunc("/api/v1/print_jobs", getPrintJobsHandler).Methods("GET")
	router.HandleFunc("/api/v1/print_jobs/{job_id}/status", updatePrintJobStatusHandler).Methods("POST")

	// Cluster management
	router.HandleFunc("/join", joinHandler).Methods("POST")
	router.HandleFunc("/remove", removeHandler).Methods("POST")

	fmt.Println("HTTP server running on port", port)
	http.ListenAndServe(":"+port, router)
}

// ----- Printers -----

func createPrinterHandler(w http.ResponseWriter, r *http.Request) {
	if raftNode.State() != raft.Leader {
		leader := raftNode.Leader()
		fmt.Printf("Redirecting to leader %s for printer creation\n", leader)
		http.Redirect(w, r, fmt.Sprintf("http://%s/api/v1/printers", leader), http.StatusTemporaryRedirect)
		return
	}

	var printer raftnode.Printer
	if err := json.NewDecoder(r.Body).Decode(&printer); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := raftnode.Command{Op: "add_printer", Printer: printer}
	data, _ := json.Marshal(cmd)

	if err := raftNode.Apply(data, 0).Error(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getPrintersHandler(w http.ResponseWriter, r *http.Request) {
	if fsm == nil {
		http.Error(w, "FSM not initialized", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(fsm.GetPrinters())
}

// ----- Filaments -----

func createFilamentHandler(w http.ResponseWriter, r *http.Request) {
	if raftNode.State() != raft.Leader {
		leader := raftNode.Leader()
		fmt.Printf("Redirecting to leader %s for filament creation\n", leader)
		http.Redirect(w, r, fmt.Sprintf("http://%s/api/v1/filaments", leader), http.StatusTemporaryRedirect)
		return
	}

	var filament raftnode.Filament
	if err := json.NewDecoder(r.Body).Decode(&filament); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := raftnode.Command{Op: "add_filament", Filament: filament}
	data, _ := json.Marshal(cmd)

	if err := raftNode.Apply(data, 0).Error(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getFilamentsHandler(w http.ResponseWriter, r *http.Request) {
	if fsm == nil {
		http.Error(w, "FSM not initialized", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(fsm.GetFilaments())
}

// ----- Print Jobs -----

func createPrintJobHandler(w http.ResponseWriter, r *http.Request) {
	if raftNode.State() != raft.Leader {
		leader := raftNode.Leader()
		fmt.Printf("Redirecting to leader %s for print job creation\n", leader)
		http.Redirect(w, r, fmt.Sprintf("http://%s/api/v1/print_jobs", leader), http.StatusTemporaryRedirect)
		return
	}

	var job raftnode.PrintJob
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	job.Status = "Queued"

	cmd := raftnode.Command{Op: "add_print_job", PrintJob: job}
	data, _ := json.Marshal(cmd)

	if err := raftNode.Apply(data, 0).Error(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getPrintJobsHandler(w http.ResponseWriter, r *http.Request) {
	if fsm == nil {
		http.Error(w, "FSM not initialized", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(fsm.GetPrintJobs())
}

func updatePrintJobStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["job_id"]
	status := r.URL.Query().Get("status")

	if jobID == "" || status == "" {
		http.Error(w, "Missing job_id or status", http.StatusBadRequest)
		return
	}

	cmd := raftnode.Command{
		Op:        "update_job_status",
		JobID:     jobID,
		NewStatus: status,
	}
	data, _ := json.Marshal(cmd)

	if err := raftNode.Apply(data, 0).Error(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ----- Cluster Management -----

type JoinRequest struct {
	ID      string `json:"id"`
	Address string `json:"address"` // Example: 127.0.0.1:5002
}

func joinHandler(w http.ResponseWriter, r *http.Request) {
	var req JoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	configFuture := raftNode.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if already part of cluster
	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == raft.ServerID(req.ID) || srv.Address == raft.ServerAddress(req.Address) {
			fmt.Fprintf(w, "Node %s at %s already joined", req.ID, req.Address)
			return
		}
	}

	// Add as Voter
	if err := raftNode.AddVoter(raft.ServerID(req.ID), raft.ServerAddress(req.Address), 0, 0).Error(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Node %s at %s joined successfully", req.ID, req.Address)
}

func removeHandler(w http.ResponseWriter, r *http.Request) {
	var req JoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := raftNode.RemoveServer(raft.ServerID(req.ID), 0, 0).Error(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Node %s removed successfully", req.ID)
}
