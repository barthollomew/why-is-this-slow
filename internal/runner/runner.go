package runner

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	"github.com/barthollomew/why-is-this-slow/internal/model"
	"github.com/barthollomew/why-is-this-slow/internal/stats"
)

const stderrLimit = 64 * 1024

type Options struct {
	Command []string
	CWD     string
	Repeat  int
}

// execute runs the command n times and captures timing and usage.
func Execute(ctx context.Context, opts Options) (model.RunResult, error) {
	if len(opts.Command) == 0 {
		return model.RunResult{}, errors.New("no command provided")
	}
	if opts.Repeat < 1 {
		opts.Repeat = 1
	}

	var samples []model.Sample
	var stderrTail string
	var exitCode int
	var signal string
	var maxRSS int64
	var maxRSSUnit string
	var wallForAggregate float64
	var cpuForAggregate float64

	cwd := opts.CWD
	if cwd == "" {
		val, err := os.Getwd()
		if err != nil {
			return model.RunResult{}, err
		}
		cwd = val
	}

	for i := 0; i < opts.Repeat; i++ {
		sample, tail, err := runOnce(ctx, opts.Command, cwd)
		if ctx.Err() != nil {
			return model.RunResult{}, ctx.Err()
		}
		if err != nil && !isExitCodeError(err) {
			return model.RunResult{}, err
		}

		samples = append(samples, sample)
		if tail != "" {
			stderrTail = tail
		}
		if sample.ExitCode != 0 {
			exitCode = sample.ExitCode
			signal = sample.Signal
		}
		if sample.MaxRSS > maxRSS {
			maxRSS = sample.MaxRSS
			maxRSSUnit = sample.MaxRSSUnit
		} else if maxRSSUnit == "" {
			maxRSSUnit = sample.MaxRSSUnit
		}
		wallForAggregate = sample.WallMS
		cpuForAggregate = sample.CPURatio
	}

	medianWall := stats.Median(getWall(samples))
	p90Wall := stats.Percentile(getWall(samples), 90)
	medianCPU := stats.Median(getCPU(samples))
	userMed := stats.Median(getUser(samples))
	sysMed := stats.Median(getSys(samples))

	run := model.RunResult{
		ID:         newRunID(),
		Timestamp:  time.Now().UTC(),
		Command:    opts.Command,
		CWD:        cwd,
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		WallMS:     medianWall,
		UserMS:     userMed,
		SysMS:      sysMed,
		CPURatio:   medianCPU,
		MaxRSSRaw:  maxRSS,
		MaxRSSUnit: maxRSSUnit,
		ExitCode:   exitCode,
		Signal:     signal,
	}

	if stderrTail != "" {
		run.StderrTail = stderrTail
	}

	if opts.Repeat > 1 {
		run.Repeat = &model.Repeat{
			Count:          opts.Repeat,
			MedianWallMS:   medianWall,
			P90WallMS:      p90Wall,
			MedianCPURatio: medianCPU,
			Samples:        samples,
		}
	}

	run.RawSamples = samples

	// single run uses the actual wall time.
	if opts.Repeat == 1 {
		run.WallMS = wallForAggregate
		run.CPURatio = cpuForAggregate
	}

	return run, nil
}

func runOnce(ctx context.Context, command []string, cwd string) (model.Sample, string, error) {
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = cwd

	tail := NewTailWriter(stderrLimit)
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, tail)

	start := time.Now()
	err := cmd.Start()
	if err != nil {
		return model.Sample{}, "", err
	}

	waitErr := cmd.Wait()
	elapsed := time.Since(start)

	usage, ok := childUsage(cmd.ProcessState)
	if !ok {
		usage.MaxRSSUnit = "unknown"
	}

	exitCode, signal := exitInfo(cmd.ProcessState, waitErr)
	wallMs := float64(elapsed) / float64(time.Millisecond)
	cpuRatio := 0.0
	if wallMs > 0 {
		cpuRatio = (usage.UserMS + usage.SysMS) / wallMs
	}

	sample := model.Sample{
		WallMS:     wallMs,
		UserMS:     usage.UserMS,
		SysMS:      usage.SysMS,
		CPURatio:   cpuRatio,
		MaxRSS:     usage.MaxRSS,
		MaxRSSUnit: usage.MaxRSSUnit,
		ExitCode:   exitCode,
		Signal:     signal,
	}

	return sample, string(tail.Bytes()), waitErr
}

func exitInfo(ps *os.ProcessState, waitErr error) (int, string) {
	exitCode := 0
	signal := ""
	if ws, ok := ps.Sys().(syscall.WaitStatus); ok {
		if ws.Signaled() {
			sigName := ws.Signal().String()
			signal = sigName
			exitCode = 128 + int(ws.Signal())
		} else {
			exitCode = ws.ExitStatus()
		}
	} else if waitErr != nil {
		exitCode = 1
	}
	return exitCode, signal
}

func isExitCodeError(err error) bool {
	var ee *exec.ExitError
	return errors.As(err, &ee)
}

func newRunID() string {
	buf := make([]byte, 4)
	_, _ = rand.Read(buf)
	suffix := hex.EncodeToString(buf)
	ts := time.Now().UTC().Format("20060102T150405Z0700")
	return fmt.Sprintf("%s-%s", ts, suffix)
}

func getWall(samples []model.Sample) []float64 {
	out := make([]float64, 0, len(samples))
	for _, s := range samples {
		out = append(out, s.WallMS)
	}
	return out
}

func getCPU(samples []model.Sample) []float64 {
	out := make([]float64, 0, len(samples))
	for _, s := range samples {
		out = append(out, s.CPURatio)
	}
	return out
}

func getUser(samples []model.Sample) []float64 {
	out := make([]float64, 0, len(samples))
	for _, s := range samples {
		out = append(out, s.UserMS)
	}
	return out
}

func getSys(samples []model.Sample) []float64 {
	out := make([]float64, 0, len(samples))
	for _, s := range samples {
		out = append(out, s.SysMS)
	}
	return out
}
