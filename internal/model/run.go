package model

import "time"

type RunResult struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Command     []string  `json:"command"`
	CWD         string    `json:"cwd"`
	Platform    string    `json:"platform"`
	WallMS      float64   `json:"wall_ms"`
	UserMS      float64   `json:"user_ms"`
	SysMS       float64   `json:"sys_ms"`
	CPURatio    float64   `json:"cpu_ratio"`
	MaxRSSRaw   int64     `json:"max_rss_raw"`
	MaxRSSUnit  string    `json:"max_rss_unit"`
	ExitCode    int       `json:"exit_code"`
	Signal      string    `json:"signal,omitempty"`
	StderrTail  string    `json:"stderr_tail,omitempty"`
	Repeat      *Repeat   `json:"repeat,omitempty"`
	RawSamples  []Sample  `json:"raw_samples,omitempty"`
	StoragePath string    `json:"-"`
}

type Repeat struct {
	Count          int      `json:"count"`
	MedianWallMS   float64  `json:"median_wall_ms"`
	P90WallMS      float64  `json:"p90_wall_ms"`
	MedianCPURatio float64  `json:"median_cpu_ratio"`
	Samples        []Sample `json:"samples"`
}

type Sample struct {
	WallMS     float64 `json:"wall_ms"`
	UserMS     float64 `json:"user_ms"`
	SysMS      float64 `json:"sys_ms"`
	CPURatio   float64 `json:"cpu_ratio"`
	MaxRSS     int64   `json:"max_rss_raw"`
	MaxRSSUnit string  `json:"max_rss_unit,omitempty"`
	ExitCode   int     `json:"exit_code"`
	Signal     string  `json:"signal,omitempty"`
}
