package store

type CommandType string

const (
	AddPrinter  CommandType = "add_printer"
	AddFilament CommandType = "add_filament"
	AddJob      CommandType = "add_job"
	UpdateJob   CommandType = "update_job"
)

const (
	Queued    = "Queued"
	Running   = "Running"
	Done      = "Done"
	Cancelled = "Cancelled"
)

type Printer struct {
	ID      string `json:"id"`
	Company string `json:"company"`
	Model   string `json:"model"`
}

type Filament struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	Color           string `json:"color"`
	TotalWeight     int    `json:"total_weight_in_grams"`
	RemainingWeight int    `json:"remaining_weight_in_grams"`
}

type PrintJob struct {
	ID         string `json:"id"`
	PrinterID  string `json:"printer_id"`
	FilamentID string `json:"filament_id"`
	Filepath   string `json:"filepath"`
	Weight     int    `json:"print_weight_in_grams"`
	Status     string `json:"status"`
}

type Command struct {
	Type    CommandType `json:"type"`
	Payload []byte      `json:"payload"`
}

type Printers struct {
	Entries []struct {
		Key   string  `json:"key"`
		Value Printer `json:"value"`
	} `json:"entries"`
}

type Filaments struct {
	Entries []struct {
		Key   string   `json:"key"`
		Value Filament `json:"value"`
	} `json:"entries"`
}

type PrintJobs struct {
	Entries []struct {
		Key   string   `json:"key"`
		Value PrintJob `json:"value"`
	} `json:"entries"`
}

