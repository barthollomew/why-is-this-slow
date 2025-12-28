# why-is-this-slow

why-is-this-slow is a small Go CLI that runs a command, records timing and resource facts, and keeps them so you can explain or compare later.

## Details
- What it is: lightweight timer with repeat mode for stable medians; captures child CPU (user+sys) time and max RSS via rusage when supported; stores every run under an OS-appropriate state directory for later explain and compare.
- What it is not: a profiler or tracer; it will not pinpoint specific functions or syscalls. Use perf, strace, or flamegraphs when you need deep attribution.
- Defaults: stdout/stderr stream live; only the last 64KB of stderr is kept in memory. Non-zero exits are recorded and returned.

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
- cpu_ratio = (user_ms + sys_ms) / wall_ms
- WAIT_IO_BOUND <0.35: likely sleeps, I/O waits, or locks.
- MIXED 0.35-0.75: combined CPU and waits.
- CPU_BOUND 0.75-1.0: mostly CPU.
- PARALLEL_CPU >1.0: CPU time exceeds wall (multi-core or multiple processes).
- Limitations: cannot tell disk vs network waits without tracing; context switches or scheduling pauses can skew ratios.

## RSS units and limits
- Linux ru_maxrss is reported in kilobytes; macOS reports bytes. Output includes the unit alongside the raw value.
- Thresholds are conservative (defaults ~512MB) and act as hints, not hard diagnostics.
- On platforms without rusage support, CPU and RSS fields fall back to zeros/unknown, but runs are still recorded.

## Troubleshooting
- State directory:
  - Linux: `$XDG_STATE_HOME/why-is-this-slow` or `~/.local/state/why-is-this-slow`
  - macOS: `~/Library/Application Support/why-is-this-slow`
  - Fallback: `~/.why-is-this-slow`
- Ensure the command to measure follows `--`; flags before `--` belong to why-is-this-slow.
- Interactive commands still stream stdout/stderr; only the last 64KB of stderr is kept in-memory for reporting.
- If CPU/RSS are zero, the platform likely lacks rusage; run on macOS/Linux for richer data.

## How it differs
- hyperfine: repeats commands and reports medians and percentiles; clean output and solid stats. Missing CPU time, RSS, persistence, and the why explanations. Great stopwatch, zero introspection. why-is-this-slow adds resource metrics, storage, and heuristic guidance.
- GNU time (/usr/bin/time -v): wall time, user+sys CPU, max RSS. Dirt simple and reliable. Missing repeat mode, storage, comparison, and interpretation. why-is-this-slow keeps the numbers, aggregates them, and explains them.
- benchstat: compares benchmark outputs statistically and excels for Go test benchmarks. Missing arbitrary commands, system metrics, and heuristics. why-is-this-slow works for any shell command and adds system insight.
- time: quick wall clock view. why-is-this-slow adds CPU ratio, RSS, storage, and explanations.
- strace/dtruss: syscall-level visibility with high detail. why-is-this-slow gives quick heuristics without tracing; pair them when you need syscall breakdowns.
- perf: deep CPU profiling, counters, and flamegraphs with kernel-level insight. Powerful but heavy. why-is-this-slow answers whether you are waiting or burning CPU with minimal overhead.
