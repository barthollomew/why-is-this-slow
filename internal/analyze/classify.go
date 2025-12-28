package analyze

import "github.com/barthollomew/why-is-this-slow/internal/model"

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

	analysis.Explanations = append(analysis.Explanations, highSysTime(run)...)
	analysis.Explanations = append(analysis.Explanations, memoryPressure(run)...)

	analysis.Notes = append(analysis.Notes, rssUnitNote(run))

	return analysis
}
