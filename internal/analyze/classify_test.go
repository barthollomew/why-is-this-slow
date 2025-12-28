package analyze

import "testing"

func TestClassifyThresholds(t *testing.T) {
	cases := []struct {
		ratio float64
		want  string
	}{
		{0.1, ClassificationWaitIO},
		{0.5, ClassificationMixed},
		{0.9, ClassificationCPU},
		{1.2, ClassificationParCPU},
	}

	for _, tc := range cases {
		if got := Classify(tc.ratio); got != tc.want {
			t.Fatalf("ratio %.2f => %s, want %s", tc.ratio, got, tc.want)
		}
	}
}
