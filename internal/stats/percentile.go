package stats

import (
	"math"
	"sort"
)

func Median(values []float64) float64 {
	return Percentile(values, 50)
}

func Percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	cp := append([]float64(nil), values...)
	sort.Float64s(cp)

	if p <= 0 {
		return cp[0]
	}
	if p >= 100 {
		return cp[len(cp)-1]
	}

	pos := (p / 100) * float64(len(cp)-1)
	lower := int(math.Floor(pos))
	upper := int(math.Ceil(pos))

	if lower == upper {
		return cp[lower]
	}

	weight := pos - float64(lower)
	return cp[lower]*(1-weight) + cp[upper]*weight
}
