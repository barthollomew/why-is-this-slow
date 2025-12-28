//go:build !linux && !darwin

package runner

import "os"

func childUsage(ps *os.ProcessState) (Usage, bool) {
	return Usage{MaxRSSUnit: "unsupported"}, false
}
