package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/barthollomew/why-is-this-slow/internal/store"
)

// Command is a lightweight subcommand wrapper.
type Command struct {
	Name    string
	Summary string
	FlagSet *flag.FlagSet
	Run     func(ctx context.Context, args []string) (int, error)
}

// Execute dispatches to subcommands. Default action is to show help.
func Execute(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	st, err := store.New()
	if err != nil {
		fmt.Fprintf(stderr, "failed to initialize storage: %v\n", err)
		return 1
	}

	cmds := []*Command{
		NewRunCommand(st, stdout),
		NewExplainCommand(st, stdout),
		NewCompareCommand(st, stdout),
	}

	index := map[string]*Command{}
	for _, c := range cmds {
		index[c.Name] = c
		c.FlagSet.SetOutput(stderr)
	}

	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		printHelp(stdout, cmds)
		return 0
	}

	cmdName := args[0]
	cmd, ok := index[cmdName]
	if !ok {
		fmt.Fprintf(stderr, "unknown subcommand %q\n\n", cmdName)
		printHelp(stderr, cmds)
		return 2
	}

	if err := cmd.FlagSet.Parse(args[1:]); err != nil {
		// flag package already prints a message to stderr.
		return 2
	}

	code, err := cmd.Run(ctx, cmd.FlagSet.Args())
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		if code != 0 {
			return code
		}
		return 1
	}

	return code
}

func printHelp(out io.Writer, cmds []*Command) {
	fmt.Fprintf(out, "why-is-this-slow - quick heuristics for slow commands\n\n")
	fmt.Fprintf(out, "Usage:\n")
	fmt.Fprintf(out, "  why-is-this-slow <command> [options]\n\n")
	fmt.Fprintf(out, "Commands:\n")

	sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name < cmds[j].Name })
	for _, c := range cmds {
		fmt.Fprintf(out, "  %-10s %s\n", c.Name, c.Summary)
	}

	fmt.Fprintf(out, "\nUse \"why-is-this-slow <command> --help\" for details.\n")
	os.Stderr.Sync()
}
