package raft

import (
	"encoding/json"
	"io"
	"sync"

	hashicraft "github.com/hashicorp/raft"
	"github.com/Miku7676/Raft3D/store"
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
	json.Unmarshal(log.Data, &cmd)

	f.Mu.Lock()
	defer f.Mu.Unlock()

	switch cmd.Type {
	case store.AddPrinter:
		var p store.Printer
		json.Unmarshal(cmd.Payload, &p)
		f.Printers[p.ID] = p

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
		var j store.PrintJob
		json.Unmarshal(cmd.Payload, &j)
		if job, ok := f.Jobs[j.ID]; ok {
			if (j.Status == store.Running && job.Status == store.Queued) ||
				(j.Status == store.Done && job.Status == store.Running) ||
				(j.Status == store.Cancelled && (job.Status == store.Queued || job.Status == store.Running)) {

				if j.Status == store.Done {
					fl := f.Filaments[j.FilamentID]
					fl.RemainingWeight -= j.Weight
					f.Filaments[j.FilamentID] = fl
				}

				job.Status = j.Status
				f.Jobs[j.ID] = job
			}
		}
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
