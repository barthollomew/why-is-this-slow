package stats

import "testing"

func TestPercentileBasic(t *testing.T) {
	vals := []float64{1, 2, 3, 4, 5}
	if got := Median(vals); got != 3 {
		t.Fatalf("median = %v", got)
	}
	if got := Percentile(vals, 90); got < 4.5 || got > 4.7 {
		t.Fatalf("p90 = %v", got)
	}
}

func TestPercentileInterpolation(t *testing.T) {
	vals := []float64{10, 20, 30, 40}
	if got := Percentile(vals, 50); got != 25 {
		t.Fatalf("median = %v", got)
	}
	if got := Percentile(vals, 25); got != 17.5 {
		t.Fatalf("p25 = %v", got)
	}
}
