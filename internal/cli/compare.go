package cli

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/barthollomew/why-is-this-slow/internal/analyze"
	"github.com/barthollomew/why-is-this-slow/internal/model"
	"github.com/barthollomew/why-is-this-slow/internal/output"
	"github.com/barthollomew/why-is-this-slow/internal/store"
)

func NewCompareCommand(st *store.Store, stdout io.Writer) *Command {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "output JSON")

	fs.Usage = func() {
		fmt.Fprintf(stdout, "Usage: why-is-this-slow compare [--json] <run_id_a> <run_id_b>\n")
		fs.PrintDefaults()
	}

	return &Command{
		Name:    "compare",
		Summary: "Compare two recorded runs",
		FlagSet: fs,
		Run: func(ctx context.Context, args []string) (int, error) {
			if len(args) < 2 {
				return 1, fmt.Errorf("two run ids are required")
			}

			runA, analysisA, err := st.Load(args[0])
			if err != nil {
				return 1, err
			}
			runB, analysisB, err := st.Load(args[1])
			if err != nil {
				return 1, err
			}

			_ = analysisA
			_ = analysisB

			compAnalysis := analyze.CompareAnalysis(runA, runB)

			if *jsonOut {
				comp := struct {
					A model.RunResult `json:"a"`
					B model.RunResult `json:"b"`
				}{
					A: runA,
					B: runB,
				}
				if err := output.WriteJSON(stdout, comp, compAnalysis); err != nil {
					return 1, err
				}
			} else {
				output.PrintCompareSummary(stdout, runA, runB, compAnalysis)
			}

			return 0, nil
		},
	}
}
