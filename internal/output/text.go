package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/barthollomew/why-is-this-slow/internal/model"
)

func PrintRunSummary(out io.Writer, run model.RunResult, analysis model.Analysis, storePath string) {
	fmt.Fprintf(out, "Command: %s\n", strings.Join(run.Command, " "))
	if run.CWD != "" {
		fmt.Fprintf(out, "CWD: %s\n", run.CWD)
	}

	if run.Repeat != nil && run.Repeat.Count > 1 {
		fmt.Fprintf(out, "Wall: median %.1fms p90 %.1fms (n=%d)\n", run.Repeat.MedianWallMS, run.Repeat.P90WallMS, run.Repeat.Count)
	} else {
		fmt.Fprintf(out, "Wall: %.1fms\n", run.WallMS)
	}

	fmt.Fprintf(out, "CPU: user %.1fms sys %.1fms cpu_ratio %.2f\n", run.UserMS, run.SysMS, run.CPURatio)
	fmt.Fprintf(out, "Max RSS: %d %s (%s)\n", run.MaxRSSRaw, safeUnit(run.MaxRSSUnit), run.Platform)
	fmt.Fprintf(out, "Exit: code=%d", run.ExitCode)
	if run.Signal != "" {
		fmt.Fprintf(out, " signal=%s", run.Signal)
	}
	fmt.Fprint(out, "\n")

	top := pickTopExplanation(analysis.Explanations)
	fmt.Fprintf(out, "Classification: %s\n", analysis.Classification)
	fmt.Fprintf(out, "Top insight: %s - %s\n", top.ID, top.Message)
	if len(top.Suggestions) > 0 {
		fmt.Fprintf(out, "Suggestions:\n")
		for i, s := range top.Suggestions {
			if i >= 4 {
				break
			}
			fmt.Fprintf(out, "  - %s\n", s)
		}
	}
	fmt.Fprintf(out, "Run ID: %s\n", run.ID)
	if storePath != "" {
		fmt.Fprintf(out, "Stored at: %s\n", storePath)
	}
}

func PrintCompareSummary(out io.Writer, a, b model.RunResult, analysis model.Analysis) {
	fmt.Fprintf(out, "Compare %s -> %s\n", a.ID, b.ID)
	fmt.Fprintf(out, "A cmd: %s\n", strings.Join(a.Command, " "))
	fmt.Fprintf(out, "B cmd: %s\n", strings.Join(b.Command, " "))
	fmt.Fprintf(out, "A: wall %.1fms cpu_ratio %.2f max_rss %d %s exit %d\n", a.WallMS, a.CPURatio, a.MaxRSSRaw, safeUnit(a.MaxRSSUnit), a.ExitCode)
	fmt.Fprintf(out, "B: wall %.1fms cpu_ratio %.2f max_rss %d %s exit %d\n", b.WallMS, b.CPURatio, b.MaxRSSRaw, safeUnit(b.MaxRSSUnit), b.ExitCode)

	top := pickTopExplanation(analysis.Explanations)
	fmt.Fprintf(out, "Classification (B): %s\n", analysis.Classification)
	fmt.Fprintf(out, "Top insight: %s - %s\n", top.ID, top.Message)
	if len(top.Suggestions) > 0 {
		fmt.Fprintf(out, "Suggestions:\n")
		for i, s := range top.Suggestions {
			if i >= 4 {
				break
			}
			fmt.Fprintf(out, "  - %s\n", s)
		}
	}
}

func pickTopExplanation(explanations []model.Explanation) model.Explanation {
	if len(explanations) == 0 {
		return model.Explanation{
			ID:          "NONE",
			Severity:    "info",
			Message:     "No explanations produced",
			Details:     "",
			Suggestions: []string{"Re-run with --repeat for more signal"},
		}
	}
	best := explanations[0]
	bestScore := severityScore(best.Severity)
	for _, e := range explanations[1:] {
		if score := severityScore(e.Severity); score > bestScore {
			best = e
			bestScore = score
		}
	}
	return best
}

func safeUnit(u string) string {
	if u == "" {
		return "units"
	}
	return u
}

func severityScore(s string) int {
	switch s {
	case "critical":
		return 3
	case "warn":
		return 2
	default:
		return 1
	}
}
