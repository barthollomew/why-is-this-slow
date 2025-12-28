//go:build linux || darwin

package runner

import (
	"os"
	"runtime"
	"syscall"
)

// childUsage extracts usage stats from a completed process using rusage.
// ru_maxrss units differ by platform: on Linux it is in kilobytes, on macOS bytes.
func childUsage(ps *os.ProcessState) (Usage, bool) {
	ru, ok := ps.SysUsage().(*syscall.Rusage)
	if !ok || ru == nil {
		return Usage{MaxRSSUnit: "unknown"}, false
	}

	userMS := float64(ru.Utime.Sec)*1000 + float64(ru.Utime.Usec)/1000
	sysMS := float64(ru.Stime.Sec)*1000 + float64(ru.Stime.Usec)/1000

	unit := "kilobytes"
	if runtime.GOOS == "darwin" {
		unit = "bytes"
	}

	return Usage{
		UserMS:     userMS,
		SysMS:      sysMS,
		MaxRSS:     ru.Maxrss,
		MaxRSSUnit: unit,
	}, true
}
