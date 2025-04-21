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
	ID              string `json:"id"`  //can be int or string but must be unique for every filament
	Type            string `json:"type"` //type:string; options: PLA, PETG, ABS, TPU
	Color           string `json:"color"` //type: string; eg: red, blue, black etc
	TotalWeight     int    `json:"total_weight_in_grams"`  //type: int
	RemainingWeight int    `json:"remaining_weight_in_grams"`  //type: int
}

type PrintJob struct {
	ID         string `json:"id"`  //can be int or string but must be unique for every print_job
	PrinterID  string `json:"printer_id"`  //needs to be a valid id of a printer that exists
	FilamentID string `json:"filament_id"` //needs to be a valid id of a filament that exists
	Filepath   string `json:"filepath"` //type: string, eg:prints/sword/hilt.gcode
	Weight     int    `json:"print_weight_in_grams"` //type: int
	Status     string `json:"status"` //type: string; options: Queued, Running, Cancelled, Done
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

