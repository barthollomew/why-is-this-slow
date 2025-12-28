package analyze

import (
	"fmt"

	"github.com/barthollomew/why-is-this-slow/internal/model"
)

const (
	ClassificationWaitIO = "WAIT_IO_BOUND"
	ClassificationMixed  = "MIXED"
	ClassificationCPU    = "CPU_BOUND"
	ClassificationParCPU = "PARALLEL_CPU"
)

func Classify(cpuRatio float64) string {
	switch {
	case cpuRatio < 0.35:
		return ClassificationWaitIO
	case cpuRatio < 0.75:
		return ClassificationMixed
	case cpuRatio <= 1.0:
		return ClassificationCPU
	default:
		return ClassificationParCPU
	}
}

// AnalyzeRun builds a model.Analysis with simple heuristics.
func AnalyzeRun(run model.RunResult) model.Analysis {
	analysis := model.Analysis{
		Classification: Classify(run.CPURatio),
	}

	analysis.Explanations = append(analysis.Explanations, baseExplanation(run, analysis.Classification))

	sysExpl := highSysTime(run)
	if len(sysExpl) > 0 {
		analysis.Notes = append(analysis.Notes, fmt.Sprintf("HIGH_SYS_TIME triggered (sys_ms/wall_ms=%.2f)", ratio(run.SysMS, run.WallMS)))
	}
	analysis.Explanations = append(analysis.Explanations, sysExpl...)

	memExpl := memoryPressure(run)
	if len(memExpl) > 0 {
		analysis.Notes = append(analysis.Notes, "MEMORY_PRESSURE triggered for single run")
	}
	analysis.Explanations = append(analysis.Explanations, memExpl...)

	analysis.Notes = append(analysis.Notes, rssUnitNote(run))

	return analysis
}

func ratio(num, denom float64) float64 {
	if denom == 0 {
		return 0
	}
	return num / denom
}
