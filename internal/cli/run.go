package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/barthollomew/why-is-this-slow/internal/analyze"
	"github.com/barthollomew/why-is-this-slow/internal/output"
	"github.com/barthollomew/why-is-this-slow/internal/runner"
	"github.com/barthollomew/why-is-this-slow/internal/store"
)

func NewRunCommand(st *store.Store, stdout io.Writer) *Command {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "output JSON")
	repeat := fs.Int("repeat", 1, "repeat N times and aggregate (median/p90)")

	fs.Usage = func() {
		fmt.Fprintf(stdout, "Usage: why-is-this-slow run [--json] [--repeat N] -- <command> [args...]\n")
		fs.PrintDefaults()
	}

	return &Command{
		Name:    "run",
		Summary: "Execute a command and record timings",
		FlagSet: fs,
		Run: func(ctx context.Context, args []string) (int, error) {
			if len(args) == 0 {
				return 1, fmt.Errorf("missing command to run; provide it after --")
			}
			if *repeat < 1 {
				return 1, fmt.Errorf("--repeat must be >=1")
			}

			res, err := runner.Execute(ctx, runner.Options{
				Command: args,
				Repeat:  *repeat,
			})
			if err != nil {
				return 1, err
			}

			analysis := analyze.AnalyzeRun(res)
			path, err := st.Save(res, analysis)
			if err != nil {
				return 1, err
			}
			res.StoragePath = path

			if *jsonOut {
				if err := output.WriteJSON(stdout, res, analysis); err != nil {
					return 1, err
				}
			} else {
				output.PrintRunSummary(stdout, res, analysis, path)
			}

			return res.ExitCode, nil
		},
	}
}

// FormatArgs rebuilds a friendly command string for display.
func FormatArgs(args []string) string {
	return strings.Join(args, " ")
}
