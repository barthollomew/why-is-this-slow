# why-is-this-slow

`why-is-this-slow` is a tiny Go CLI that runs a command, captures wall time, child CPU usage, and max RSS, then stores the results for later explanation or comparison. It is intentionally heuristic and fast: no tracing, no perf sampling, just honest numbers and simple rules of thumb.

## What it is / what it isn't
- A lightweight timer with a repeat mode for stable medians.
- Captures child CPU (user+sys) time and max RSS via `rusage` when supported.
- Stores every run under an OS-appropriate state directory for later `explain` and `compare`.
- Not a profiler or tracer; it will not pinpoint specific functions or syscalls.
- Not a replacement for `perf`, `strace`, or flamegraphs—use those when you need precision.

## Install
- Go toolchain required.
- `go install github.com/barthollomew/why-is-this-slow/cmd/why-is-this-slow@latest`
- Build from source: `go build -o why-is-this-slow ./cmd/why-is-this-slow`

## Usage
```
why-is-this-slow run [--json] [--repeat N] -- <command> [args...]
why-is-this-slow explain [--json] <run_id>
why-is-this-slow compare [--json] <run_id_a> <run_id_b>
```

### Quickstart
- Run once: `why-is-this-slow run -- ls -l`
- Repeat for stability: `why-is-this-slow run --repeat 3 -- sleep 0.1`
- Inspect later: `why-is-this-slow explain <run_id>`
- Compare two captures: `why-is-this-slow compare <id_a> <id_b>`
- JSON output for scripting: add `--json` to any subcommand.

Example text output:
```
Command: sleep 0.1
Wall: median 101.2ms p90 105.4ms (n=3)
CPU: user 0.0ms sys 0.0ms cpu_ratio 0.00
Max RSS: 0 unknown (linux/amd64)
Exit: code=0
Classification: WAIT_IO_BOUND
Top insight: BASELINE - Likely waiting on I/O or sleeps
Suggestions:
  - Check if the command is expected to wait on I/O or locks
  - Trim unnecessary work or add tracing if unsure
Run ID: 20240101T120000Z-1a2b3c4d
Stored at: ~/.local/state/why-is-this-slow/runs/20240101T120000Z-1a2b3c4d.json
```

## Interpreting cpu_ratio
- `cpu_ratio = (user_ms + sys_ms) / wall_ms`
- WAIT_IO_BOUND `<0.35`: likely sleeps, I/O waits, or locks.
- MIXED `0.35-0.75`: combined CPU and waits.
- CPU_BOUND `0.75-1.0`: mostly CPU.
- PARALLEL_CPU `>1.0`: CPU time exceeds wall (multi-core or multiple processes).
- Limitations: cannot tell disk vs network waits without tracing; context switches or scheduling pauses can skew ratios.

## RSS units and limits
- Linux `ru_maxrss` is reported in kilobytes; macOS reports bytes. Output includes the unit alongside the raw value.
- Thresholds are conservative (defaults ~512MB) and act as hints, not hard diagnostics.
- On platforms without `rusage` support, CPU and RSS fields fall back to zeros/unknown, but runs are still recorded.

## Troubleshooting
- State directory:
  - Linux: `$XDG_STATE_HOME/why-is-this-slow` or `~/.local/state/why-is-this-slow`
  - macOS: `~/Library/Application Support/why-is-this-slow`
  - Fallback: `~/.why-is-this-slow`
- Ensure the command to measure follows `--`; flags before `--` belong to `why-is-this-slow`.
- Interactive commands still stream stdout/stderr; only the last 64KB of stderr is kept in-memory for reporting.
- Non-zero exits are recorded; the CLI returns the same exit code.
- If CPU/RSS are zero, the platform likely lacks `rusage`; consider running on macOS/Linux for richer data.

## Design philosophy
- Fast: minimal overhead, no heavy dependencies.
- Heuristic: honest rules of thumb, not magic.
- Transparent: versioned JSON records and explicit units.

## How it differs from time, perf, strace
- `time`: great for quick wall time; this adds CPU ratio, RSS, storage, and explanations.
- `perf`: much deeper CPU profiling; use it when you need function-level attribution.
- `strace`/`dtruss`: syscall-level visibility; pair it with this tool when system time is high.
