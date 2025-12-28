package analyze

import (
	"fmt"
	"runtime"

	"github.com/barthollomew/why-is-this-slow/internal/model"
)

func baseExplanation(run model.RunResult, classification string) model.Explanation {
	msg := "Balanced workload"
	details := fmt.Sprintf("cpu_ratio=%.2f", run.CPURatio)
	switch classification {
	case ClassificationWaitIO:
		msg = "Likely waiting on I/O or sleeps"
	case ClassificationMixed:
		msg = "Mixed CPU and waiting"
	case ClassificationCPU:
		msg = "Mostly CPU-bound"
	case ClassificationParCPU:
		msg = "CPU time exceeds wall (parallel or multi-process)"
	}

	return model.Explanation{
		ID:       "BASELINE",
		Severity: "info",
		Message:  msg,
		Details:  details,
		Suggestions: []string{
			"Check if the command is expected to wait on I/O or locks",
			"Trim unnecessary work or add tracing if unsure",
		},
	}
}

func highSysTime(run model.RunResult) []model.Explanation {
	wall := run.WallMS
	if wall <= 0 {
		return nil
	}
	ratio := run.SysMS / wall
	if ratio <= 0.25 {
		return nil
	}

	return []model.Explanation{
		{
			ID:       "HIGH_SYS_TIME",
			Severity: "warn",
			Message:  "High system time relative to wall clock",
			Details:  fmt.Sprintf("sys_ms=%.1f wall_ms=%.1f", run.SysMS, wall),
			Suggestions: []string{
				"Inspect disk or network I/O, or frequent context switches",
				"Use strace/dtruss to see syscalls if precision is needed",
			},
		},
	}
}

func memoryPressure(run model.RunResult) []model.Explanation {
	threshold := memoryThreshold()
	if threshold <= 0 {
		return nil
	}
	if run.MaxRSSRaw <= threshold {
		return nil
	}

	unit := run.MaxRSSUnit
	if unit == "" {
		unit = "units"
	}

	return []model.Explanation{
		{
			ID:       "MEMORY_PRESSURE",
			Severity: "warn",
			Message:  "Memory usage seems high",
			Details:  fmt.Sprintf("max_rss=%d (%s)", run.MaxRSSRaw, unit),
			Suggestions: []string{
				"Inspect allocations or load size; reduce concurrency if unbounded",
				"Use pprof or heap profiling for precise attribution",
			},
		},
	}
}

func memoryThreshold() int64 {
	switch runtime.GOOS {
	case "darwin":
		return 512 * 1024 * 1024 // bytes
	case "linux":
		return 512 * 1024 // kilobytes
	default:
		return 0
	}
}

func rssUnitNote(run model.RunResult) string {
	unit := run.MaxRSSUnit
	if unit == "" {
		unit = "unknown"
	}
	return fmt.Sprintf("max_rss unit for %s: %s", runtime.GOOS, unit)
}

// CompareAnalysis generates explanations based on two runs.
func CompareAnalysis(a, b model.RunResult) model.Analysis {
	analysis := model.Analysis{
		Classification: Classify(b.CPURatio),
	}
	analysis.Explanations = append(analysis.Explanations, model.Explanation{
		ID:       "COMPARISON_BASE",
		Severity: "info",
		Message:  "Comparing run B against run A",
		Details:  fmt.Sprintf("A=%s B=%s", a.ID, b.ID),
		Suggestions: []string{
			"Focus on regressions in wall time and memory first",
			"Re-run with tracing if differences are unexpected",
		},
	})

	analysis.Explanations = append(analysis.Explanations, compareMemory(a, b)...)
	analysis.Explanations = append(analysis.Explanations, compareCPU(a, b)...)
	analysis.Explanations = append(analysis.Explanations, memoryPressure(b)...)
	analysis.Explanations = append(analysis.Explanations, highSysTime(b)...)
	analysis.Notes = append(analysis.Notes, rssUnitNote(b))

	return analysis
}

func compareMemory(a, b model.RunResult) []model.Explanation {
	if a.MaxRSSRaw == 0 || b.MaxRSSRaw == 0 {
		return nil
	}
	diff := float64(b.MaxRSSRaw-a.MaxRSSRaw) / float64(a.MaxRSSRaw)
	if diff <= 0.30 {
		return nil
	}

	return []model.Explanation{
		{
			ID:       "MEMORY_PRESSURE",
			Severity: "warn",
			Message:  "Memory usage increased notably",
			Details:  fmt.Sprintf("run_b max_rss %d vs run_a %d", b.MaxRSSRaw, a.MaxRSSRaw),
			Suggestions: []string{
				"Check for new caches or data growth between runs",
				"Profile allocations or compare inputs to explain the jump",
			},
		},
	}
}

func compareCPU(a, b model.RunResult) []model.Explanation {
	if a.CPURatio == 0 {
		return nil
	}
	diff := b.CPURatio - a.CPURatio
	if diff <= 0.15 {
		return nil
	}
	return []model.Explanation{
		{
			ID:       "CPU_SHIFT",
			Severity: "info",
			Message:  "CPU utilization changed",
			Details:  fmt.Sprintf("cpu_ratio a=%.2f b=%.2f", a.CPURatio, b.CPURatio),
			Suggestions: []string{
				"Confirm expected workload; if not, inspect CPU-heavy sections",
			},
		},
	}
}
