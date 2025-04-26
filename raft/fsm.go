package raftnode

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/hashicorp/raft"
)

type Printer struct {
	ID      string `json:"id"`
	Company string `json:"company"`
	Model   string `json:"model"`
}

type Filament struct {
	ID                     string `json:"id"`
	Type                   string `json:"type"`
	Color                  string `json:"color"`
	TotalWeightInGrams     int    `json:"total_weight_in_grams"`
	RemainingWeightInGrams int    `json:"remaining_weight_in_grams"`
}

type PrintJob struct {
	ID                 string `json:"id"`
	PrinterID          string `json:"printer_id"`
	FilamentID         string `json:"filament_id"`
	FilePath           string `json:"filepath"`
	PrintWeightInGrams int    `json:"print_weight_in_grams"`
	Status             string `json:"status"` // Queued, Running, Done, Cancelled
}

type Command struct {
	Op        string   `json:"op"`
	Printer   Printer  `json:"printer,omitempty"`
	Filament  Filament `json:"filament,omitempty"`
	PrintJob  PrintJob `json:"print_job,omitempty"`
	JobID     string   `json:"job_id,omitempty"`
	NewStatus string   `json:"new_status,omitempty"`
}

type PrinterFSM struct {
	mu        sync.Mutex
	printers  map[string]Printer
	filaments map[string]Filament
	jobs      map[string]PrintJob
}

func NewPrinterFSM() *PrinterFSM {
	return &PrinterFSM{
		printers:  make(map[string]Printer),
		filaments: make(map[string]Filament),
		jobs:      make(map[string]PrintJob),
	}
}

func (p *PrinterFSM) Apply(log *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	switch cmd.Op {
	case "add_printer":
		p.printers[cmd.Printer.ID] = cmd.Printer

	case "add_filament":
		p.filaments[cmd.Filament.ID] = cmd.Filament

	case "add_print_job":
		cmd.PrintJob.Status = "Queued" // default status
		p.jobs[cmd.PrintJob.ID] = cmd.PrintJob

	case "update_job_status":
		job, exists := p.jobs[cmd.JobID]
		if !exists {
			return fmt.Errorf("job not found")
		}

		status := strings.ToLower(cmd.NewStatus)
		current := strings.ToLower(job.Status)

		switch status {
		case "running":
			if current != "queued" {
				return fmt.Errorf("can only move to running from queued")
			}
			job.Status = "Running"

		case "done":
			if current != "running" {
				return fmt.Errorf("can only mark done from running")
			}
			filament, ok := p.filaments[job.FilamentID]
			if !ok {
				return fmt.Errorf("filament not found")
			}
			if filament.RemainingWeightInGrams < job.PrintWeightInGrams {
				return fmt.Errorf("not enough filament remaining")
			}
			filament.RemainingWeightInGrams -= job.PrintWeightInGrams
			p.filaments[job.FilamentID] = filament
			job.Status = "Done"

		case "cancelled":
			if current != "queued" && current != "running" {
				return fmt.Errorf("can only cancel queued or running jobs")
			}
			job.Status = "Cancelled"

		default:
			return fmt.Errorf("invalid status transition")
		}

		p.jobs[cmd.JobID] = job
	}

	return nil
}

func (p *PrinterFSM) Snapshot() (raft.FSMSnapshot, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	state := map[string]interface{}{
		"printers":  p.printers,
		"filaments": p.filaments,
		"jobs":      p.jobs,
	}

	data, _ := json.Marshal(state)
	return &printerSnapshot{state: data}, nil
}

func (p *PrinterFSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()
	var state map[string]interface{}
	if err := json.NewDecoder(rc).Decode(&state); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	printersJSON, _ := json.Marshal(state["printers"])
	filamentsJSON, _ := json.Marshal(state["filaments"])
	jobsJSON, _ := json.Marshal(state["jobs"])

	json.Unmarshal(printersJSON, &p.printers)
	json.Unmarshal(filamentsJSON, &p.filaments)
	json.Unmarshal(jobsJSON, &p.jobs)

	return nil
}

type printerSnapshot struct {
	state []byte
}

func (s *printerSnapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := sink.Write(s.state); err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *printerSnapshot) Release() {}

func (p *PrinterFSM) GetPrinters() []Printer {
	p.mu.Lock()
	defer p.mu.Unlock()

	printers := make([]Printer, 0, len(p.printers))
	for _, printer := range p.printers {
		printers = append(printers, printer)
	}
	return printers
}

func (p *PrinterFSM) GetFilaments() []Filament {
	p.mu.Lock()
	defer p.mu.Unlock()

	filaments := make([]Filament, 0, len(p.filaments))
	for _, filament := range p.filaments {
		filaments = append(filaments, filament)
	}
	return filaments
}

func (p *PrinterFSM) GetPrintJobs() []PrintJob {
	p.mu.Lock()
	defer p.mu.Unlock()

	jobs := make([]PrintJob, 0, len(p.jobs))
	for _, job := range p.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}
