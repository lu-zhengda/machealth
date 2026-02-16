package health

import (
	"sync"
	"time"
)

// Check runs all health checks in parallel and returns a complete report.
func Check() Report {
	r := Report{
		Timestamp: time.Now().UTC(),
	}

	var wg sync.WaitGroup
	wg.Add(8)

	go func() { defer wg.Done(); r.CPU = CheckCPU() }()
	go func() { defer wg.Done(); r.Memory = CheckMemory() }()
	go func() { defer wg.Done(); r.Disk = CheckDisk() }()
	go func() { defer wg.Done(); r.Thermal = CheckThermal() }()
	go func() { defer wg.Done(); r.ICloud = CheckICloud() }()
	go func() { defer wg.Done(); r.Battery = CheckBattery() }()
	go func() { defer wg.Done(); r.TimeMachine = CheckTimeMachine() }()
	go func() { defer wg.Done(); r.Network = CheckNetwork() }()

	wg.Wait()

	r.Score = computeScore(r)
	return r
}

// Diagnose runs all checks and generates diagnoses for non-green subsystems.
func Diagnose() DiagnoseReport {
	r := Check()
	dr := DiagnoseReport{Report: r}

	diagnosers := []func() *Diagnosis{
		func() *Diagnosis { return DiagnoseCPU(r.CPU) },
		func() *Diagnosis { return DiagnoseMemory(r.Memory) },
		func() *Diagnosis { return DiagnoseDisk(r.Disk) },
		func() *Diagnosis { return DiagnoseThermal(r.Thermal) },
		func() *Diagnosis { return DiagnoseICloud(r.ICloud) },
		func() *Diagnosis { return DiagnoseBattery(r.Battery) },
		func() *Diagnosis { return DiagnoseTimeMachine(r.TimeMachine) },
		func() *Diagnosis { return DiagnoseNetwork(r.Network) },
	}

	for _, fn := range diagnosers {
		if d := fn(); d != nil {
			dr.Diagnoses = append(dr.Diagnoses, *d)
		}
	}

	if dr.Diagnoses == nil {
		dr.Diagnoses = []Diagnosis{}
	}

	return dr
}

// Subsystem weights for composite score.
var weights = map[string]int{
	"cpu":     20,
	"memory":  25,
	"disk":    15,
	"thermal": 20,
	"battery": 10,
	"icloud":  5,
	"network": 5,
}

func statusValue(s Status) int {
	switch s {
	case StatusGreen:
		return 100
	case StatusYellow:
		return 50
	default:
		return 0
	}
}

func computeScore(r Report) Score {
	subsystems := map[string]Status{
		"cpu":         r.CPU.Status,
		"memory":      r.Memory.Status,
		"disk":        r.Disk.Status,
		"thermal":     r.Thermal.Status,
		"icloud":      r.ICloud.Status,
		"battery":     r.Battery.Status,
		"timemachine": r.TimeMachine.Status,
		"network":     r.Network.Status,
	}

	// TimeMachine shares disk weight (not a separate weight)
	// but still contributes to overall status
	totalWeight := 0
	weightedSum := 0
	worstStatus := StatusGreen
	var reasons []string

	for name, status := range subsystems {
		w, ok := weights[name]
		if !ok {
			// timemachine doesn't have its own weight — skip for scoring
			// but still check for status degradation
			if status != StatusGreen {
				if worstStatus == StatusGreen || (worstStatus == StatusYellow && status == StatusRed) {
					worstStatus = status
				}
				reasons = append(reasons, name+":"+string(status))
			}
			continue
		}

		totalWeight += w
		weightedSum += w * statusValue(status)

		if status != StatusGreen {
			if worstStatus == StatusGreen || (worstStatus == StatusYellow && status == StatusRed) {
				worstStatus = status
			}
			reasons = append(reasons, name+":"+string(status))
		}
	}

	value := 0
	if totalWeight > 0 {
		value = weightedSum / totalWeight
	}

	if reasons == nil {
		reasons = []string{}
	}

	return Score{
		Status:  worstStatus,
		Value:   value,
		Reasons: reasons,
	}
}

// ExitCode returns the appropriate exit code for a report.
func ExitCode(r Report) int {
	switch r.Score.Status {
	case StatusRed:
		return 2
	case StatusYellow:
		return 1
	default:
		return 0
	}
}
