package raft

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/Miku7676/Raft3D/store"
	hashicraft "github.com/hashicorp/raft"
)

type FSM struct {
	Mu        sync.Mutex
	Printers  map[string]store.Printer
	Filaments map[string]store.Filament
	Jobs      map[string]store.PrintJob
}

func NewFSM() *FSM {
	return &FSM{
		Printers:  make(map[string]store.Printer),
		Filaments: make(map[string]store.Filament),
		Jobs:      make(map[string]store.PrintJob),
	}
}

func (f *FSM) Apply(log *hashicraft.Log) interface{} {
	var cmd store.Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		fmt.Printf("Error unmarshaling command: %v\n", err)
		return err
	}

	f.Mu.Lock()
	defer f.Mu.Unlock()

	fmt.Printf("Applying command type: %s\n", string(cmd.Type))

	switch cmd.Type {
	case store.AddPrinter:
		var p store.Printer
		if err := json.Unmarshal(cmd.Payload, &p); err != nil {
			return err
		}
		fmt.Printf("Adding printer: %+v with ID: %s\n", p, p.ID)
		if f.Printers[p.ID] != (store.Printer{}) {
			return fmt.Errorf("printer %s already exists", p.ID)
		}
		f.Printers[p.ID] = p
		fmt.Printf("Added printer %s\n%v", p.ID, len(f.Printers))

	case store.AddFilament:
		var fl store.Filament
		json.Unmarshal(cmd.Payload, &fl)
		f.Filaments[fl.ID] = fl

	case store.AddJob:
		var j store.PrintJob
		json.Unmarshal(cmd.Payload, &j)
		j.Status = store.Queued
		f.Jobs[j.ID] = j

	case store.UpdateJob:
		var updateJob store.PrintJob
		if err := json.Unmarshal(cmd.Payload, &updateJob); err != nil {
			fmt.Printf("Error in job update: %v\n", err)
			return err
		}

		existingJob, ok := f.Jobs[updateJob.ID]
		if !ok {
			fmt.Printf("Job %s not found for update\n", updateJob.ID)
			return fmt.Errorf("job %s not found", updateJob.ID)
		}


		// if job, ok := f.Jobs[j.ID]; ok {
		// 	if (j.Status == store.Running && job.Status == store.Queued) ||
		// 		(j.Status == store.Done && job.Status == store.Running) ||
		// 		(j.Status == store.Cancelled && (job.Status == store.Queued || job.Status == store.Running)) {
		//
		// 		if j.Status == store.Done {
		// 			fl := f.Filaments[j.FilamentID]
		// 			fl.RemainingWeight -= j.Weight
		// 			f.Filaments[j.FilamentID] = fl
		// 		}
		//
		// 		job.Status = j.Status
		// 		f.Jobs[j.ID] = job
		// 	}
		// }

		validTransition := false

		switch updateJob.Status {
		case store.Running:
			validTransition = existingJob.Status == store.Queued
		case store.Done:
			validTransition = existingJob.Status == store.Running
		case store.Cancelled:
			validTransition = existingJob.Status == store.Queued || existingJob.Status == store.Running
		default:
			validTransition = false
		}

		if !validTransition {
			fmt.Printf("Invalid status transition from %s to %s\n", existingJob.Status, updateJob.Status)
			return fmt.Errorf("invalid status transition from %s to %s", existingJob.Status, updateJob.Status)
		}

		// Update job status
		existingJob.Status = updateJob.Status

		// If job is done, update filament weight
		if updateJob.Status == store.Done {
			filament, filamentExists := f.Filaments[existingJob.FilamentID]
			if filamentExists {
				fmt.Printf("Updating filament %s weight from %d to %d\n", 
					existingJob.FilamentID, 
					filament.RemainingWeight, 
					filament.RemainingWeight - existingJob.Weight)

				filament.RemainingWeight -= existingJob.Weight
				f.Filaments[existingJob.FilamentID] = filament
			} else {
				fmt.Printf("Filament %s not found\n", existingJob.FilamentID)
			}
		}

		// Save updated job
		f.Jobs[updateJob.ID] = existingJob

		fmt.Printf("Job %s updated to status %s\n", existingJob.ID, existingJob.Status)
	}
	return nil
}

func (f *FSM) Snapshot() (hashicraft.FSMSnapshot, error) {
	return &snapshot{state: f}, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	decoder := json.NewDecoder(rc)
	return decoder.Decode(f)
}

type snapshot struct {
	state *FSM
}

func (s *snapshot) Persist(sink hashicraft.SnapshotSink) error {
	b, err := json.Marshal(s.state)
	if err != nil {
		sink.Cancel()
		return err
	}
	if _, err := sink.Write(b); err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *snapshot) Release() {}
