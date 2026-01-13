package analyze

import (
	"testing"

	"github.com/barthollomew/why-is-this-slow/internal/model"
)

func TestHighSysTimeRule(t *testing.T) {
	run := model.RunResult{
		WallMS: 1000,
		SysMS:  300,
	}
	expl := highSysTime(run)
	if len(expl) == 0 {
		t.Fatalf("expected high sys time explanation")
	}
}

func TestIOWaitRule(t *testing.T) {
	run := model.RunResult{
		WallMS:   1000,
		UserMS:   10,
		SysMS:    10,
		CPURatio: 0.02,
		Platform: "linux/amd64",
	}
	expl := ioWait(run)
	if len(expl) == 0 {
		t.Fatalf("expected io wait explanation")
	}
}

func TestMemoryPressureRule(t *testing.T) {
	thr := memoryThreshold()
	if thr == 0 {
		t.Skip("memory threshold unsupported on this platform")
	}
	run := model.RunResult{
		MaxRSSRaw:  thr + 1,
		MaxRSSUnit: "units",
	}
	expl := memoryPressure(run)
	if len(expl) == 0 {
		t.Fatalf("expected memory pressure explanation for threshold %d", thr)
	}
}

func TestCompareMemoryIncrease(t *testing.T) {
	a := model.RunResult{ID: "a", MaxRSSRaw: 100}
	b := model.RunResult{ID: "b", MaxRSSRaw: 140}
	expl := compareMemory(a, b)
	if len(expl) == 0 {
		t.Fatalf("expected memory increase explanation")
	}
}

func TestCompareWallRegression(t *testing.T) {
	a := model.RunResult{ID: "a", WallMS: 100}
	b := model.RunResult{ID: "b", WallMS: 140}
	expl := compareWall(a, b)
	if len(expl) == 0 {
		t.Fatalf("expected wall regression explanation")
	}
}
