# Why is this slow?

A small Go CLI that runs a command, records how long it took and what the system did, and saves the result so you can look at it later or compare runs.

This exists for the very common situation where something feels slow and you want evidence before guessing.

## What it is (and isn’t)

- It is:
  - A lightweight runner with optional repeat mode for more stable numbers.
  - Captures wall time, child CPU time (user + sys), and max RSS via `rusage` where available.
  - Stores every run under a per-OS state directory so you can explain or compare later.
  - Compares runs to flag wall-time regressions and resource shifts.

- It is not:
  - A profiler or tracer.
  - It will not tell you which function or syscall is slow.
  - If you need that, use perf, strace, dtruss, or flamegraphs.

Defaults are boring on purpose:
- stdout and stderr stream live.
- Only the last 64KB of stderr is kept in memory.
- Non-zero exits are recorded and returned.

## Install

- Requires Go.
- Install:
  ```sh
  go install github.com/barthollomew/why-is-this-slow/cmd/why-is-this-slow@latest
  ```
- Or build from source:
  ```sh
  go build -o why-is-this-slow ./cmd/why-is-this-slow
  ```

## Usage

```
why-is-this-slow run [--json] [--repeat N] -- <command> [args...]
why-is-this-slow explain [--json] <run_id>
why-is-this-slow compare [--json] <run_id_a> <run_id_b>
```

### Quickstart

- Run once:
  ```sh
  why-is-this-slow run -- ls -l
  ```
- Repeat for stability:
  ```sh
  why-is-this-slow run --repeat 3 -- sleep 0.1
  ```
- Inspect later:
  ```sh
  why-is-this-slow explain <run_id>
  ```
- Compare two runs:
  ```sh
  why-is-this-slow compare <id_a> <id_b>
  ```
- Add `--json` to any command for machine-readable output.

Example output:
```
Command: sleep 0.1
Wall: median 101.2ms p90 105.4ms (n=3)
CPU: user 0.0ms sys 0.0ms cpu_ratio 0.00
Max RSS: 0 unknown (linux/amd64)
Exit: code=0
Classification: WAIT_IO_BOUND
Top insight: HIGH_IO_WAIT - High wait time (~100% of wall)
Suggestions:
  - Check for disk/network latency, lock contention, or sleeps
  - Use tracing (strace/dtruss) if the wait is unexpected
Run ID: 20240101T120000Z-1a2b3c4d
Stored at: ~/.local/state/why-is-this-slow/runs/20240101T120000Z-1a2b3c4d.json
```

## Interpreting `cpu_ratio`

- `cpu_ratio = (user_ms + sys_ms) / wall_ms`
- `< 0.35` WAIT_IO_BOUND: mostly sleeping, waiting on I/O, or blocked on locks.
- `0.35–0.75` MIXED: some CPU, some waiting.
- `0.75–1.0` CPU_BOUND: mostly burning CPU.
- `> 1.0` PARALLEL_CPU: CPU time exceeds wall time (multiple cores or processes).

Caveats:
- Cannot distinguish disk vs network waits without tracing.
- Scheduling delays and context switches can skew ratios.

## RSS units and limits

- Linux reports `ru_maxrss` in kilobytes.
- macOS reports bytes.
- The output includes units so you do not have to remember this.
- Thresholds are conservative and meant as hints, not alarms.
- On platforms without `rusage`, CPU and RSS may be zero or unknown, but runs are still recorded.

## Troubleshooting

- State directory:
  - Linux: `$XDG_STATE_HOME/why-is-this-slow` or `~/.local/state/why-is-this-slow`
  - macOS: `~/Library/Application Support/why-is-this-slow`
  - Fallback: `~/.why-is-this-slow`
- Anything after `--` is the command being measured.
- Interactive programs still stream output normally.
- Zero CPU or RSS usually means the platform does not expose `rusage`.

## How it compares

- hyperfine:
  - Excellent timing stats.
  - No CPU time, RSS, storage, or interpretation.
- GNU time (`/usr/bin/time -v`):
  - Reliable raw numbers.
  - No repeat mode, persistence, or comparisons.
- benchstat:
  - Great for Go benchmarks.
  - Not for arbitrary shell commands or system metrics.
- strace / dtruss:
  - Deep syscall detail.
  - Heavyweight and noisy.
- perf:
  - Superr powerful.
  - Overkill when you just want to know if you are waiting or burning CPU.

why-is-this-slow sits in the gap between “stopwatch” and “profiler” and tries to answer the first useful question: what kind of slow is this?
