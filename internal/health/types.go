package health

import "time"

// Status represents the health status of a subsystem.
type Status string

const (
	StatusGreen  Status = "green"
	StatusYellow Status = "yellow"
	StatusRed    Status = "red"
)

// Report is the top-level health check output.
type Report struct {
	Timestamp   time.Time   `json:"timestamp"`
	Score       Score       `json:"score"`
	CPU         CPU         `json:"cpu"`
	Memory      Memory      `json:"memory"`
	Disk        Disk        `json:"disk"`
	Thermal     Thermal     `json:"thermal"`
	ICloud      ICloud      `json:"icloud"`
	Battery     Battery     `json:"battery"`
	TimeMachine TimeMachine `json:"timemachine"`
	Network     Network     `json:"network"`
	Bluetooth   Bluetooth   `json:"bluetooth"`
}

// Score is the composite health score.
type Score struct {
	Status  Status   `json:"status"`
	Value   int      `json:"value"`
	Reasons []string `json:"reasons"`
}

// CPU contains CPU load information.
type CPU struct {
	Status       Status  `json:"status"`
	Error        string  `json:"error,omitempty"`
	LoadAvg1m    float64 `json:"load_avg_1m"`
	LoadAvg5m    float64 `json:"load_avg_5m"`
	LoadAvg15m   float64 `json:"load_avg_15m"`
	LogicalCores int     `json:"logical_cores"`
	LoadPerCore  float64 `json:"load_per_core"`
}

// Memory contains memory pressure information.
type Memory struct {
	Status          Status  `json:"status"`
	Error           string  `json:"error,omitempty"`
	PressurePercent int     `json:"pressure_percent"`
	SwapUsedMB      float64 `json:"swap_used_mb"`
	SwapTotalMB     float64 `json:"swap_total_mb"`
}

// Disk contains disk space information.
type Disk struct {
	Status      Status  `json:"status"`
	Error       string  `json:"error,omitempty"`
	AvailableGB float64 `json:"available_gb"`
	TotalGB     float64 `json:"total_gb"`
	UsedPercent float64 `json:"used_percent"`
}

// Thermal contains thermal throttling information.
type Thermal struct {
	Status        Status `json:"status"`
	Error         string `json:"error,omitempty"`
	CPUSpeedLimit int    `json:"cpu_speed_limit"`
	Throttled     bool   `json:"throttled"`
}

// ICloud contains iCloud sync state.
type ICloud struct {
	Status   Status `json:"status"`
	Error    string `json:"error,omitempty"`
	Syncing  bool   `json:"syncing"`
	CaughtUp bool   `json:"caught_up"`
	LastSync string `json:"last_sync"`
}

// Battery contains battery state.
type Battery struct {
	Status           Status  `json:"status"`
	Error            string  `json:"error,omitempty"`
	Percent          int     `json:"percent"`
	PowerSource      string  `json:"power_source"`
	Charging         bool    `json:"charging"`
	FullyCharged     bool    `json:"fully_charged"`
	TimeRemainingMin int     `json:"time_remaining_min"`
	HealthPercent    float64 `json:"health_percent"`
	CycleCount       int     `json:"cycle_count"`
	Installed        bool    `json:"installed"`
}

// TimeMachine contains Time Machine backup state.
type TimeMachine struct {
	Status  Status  `json:"status"`
	Error   string  `json:"error,omitempty"`
	Running bool    `json:"running"`
	Phase   string  `json:"phase,omitempty"`
	Percent float64 `json:"percent"`
}

// Network contains network connectivity state.
type Network struct {
	Status    Status `json:"status"`
	Error     string `json:"error,omitempty"`
	Reachable bool   `json:"reachable"`
	Interface string `json:"interface,omitempty"`
	IP        string `json:"ip,omitempty"`
}

// Bluetooth contains Bluetooth controller and connected-device state.
type Bluetooth struct {
	Status               Status            `json:"status"`
	Error                string            `json:"error,omitempty"`
	Available            bool              `json:"available"`
	Enabled              bool              `json:"enabled"`
	ConnectedDeviceCount int               `json:"connected_device_count"`
	Devices              []BluetoothDevice `json:"devices,omitempty"`
}

// BluetoothDevice represents a single paired/connected Bluetooth device.
type BluetoothDevice struct {
	Name           string `json:"name"`
	Connected      bool   `json:"connected"`
	BatteryPercent int    `json:"battery_percent,omitempty"` // -1 means not reported
}

// Diagnosis is a detailed explanation of a health issue.
type Diagnosis struct {
	Subsystem string `json:"subsystem"`
	Severity  Status `json:"severity"`
	Summary   string `json:"summary"`
	Detail    string `json:"detail"`
	Action    string `json:"action"`
}

// DiagnoseReport extends Report with diagnoses.
type DiagnoseReport struct {
	Report
	Diagnoses []Diagnosis `json:"diagnoses"`
}
