package cli

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/barthollomew/why-is-this-slow/internal/output"
	"github.com/barthollomew/why-is-this-slow/internal/store"
)

func NewExplainCommand(st *store.Store, stdout io.Writer) *Command {
	fs := flag.NewFlagSet("explain", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "output JSON")

	fs.Usage = func() {
		fmt.Fprintf(stdout, "Usage: why-is-this-slow explain [--json] <run_id>\n")
		fs.PrintDefaults()
	}

	return &Command{
		Name:    "explain",
		Summary: "Show details for a recorded run",
		FlagSet: fs,
		Run: func(ctx context.Context, args []string) (int, error) {
			if len(args) < 1 {
				return 1, fmt.Errorf("run_id is required")
			}
			run, analysis, err := st.Load(args[0])
			if err != nil {
				return 1, err
			}

			if *jsonOut {
				if err := output.WriteJSON(stdout, run, analysis); err != nil {
					return 1, err
				}
			} else {
				output.PrintRunSummary(stdout, run, analysis, run.StoragePath)
			}
			return 0, nil
		},
	}
}
