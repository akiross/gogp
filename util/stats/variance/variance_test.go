package variance

import "testing"
import "math"

func equalFloats(a, b float64) bool {
	return math.Abs(a-b) < 0.0000001 // This comparison sucks, but here's ok
}

func TestVariance(t *testing.T) {

	tests := []struct {
		vals            []float64
		expMean, expVar float64
	}{
		{[]float64{1, 2, 3}, 2, 2 / 3.0},
		{[]float64{600, 470, 170, 430, 300}, 394, 21704},
	}

	v := Create()

	for _, test := range tests {
		for _, val := range test.vals {
			v.Accumulate(val)
		}
		if !equalFloats(v.PartialMean(), test.expMean) {
			t.Errorf("Expected mean %v, got %v", test.expMean, v.PartialMean())
		}
		if !equalFloats(v.PartialVar(), test.expVar) {
			t.Errorf("Expected var %v, got %v", test.expVar, v.PartialVar())
		}
		v.Reset()
	}
}
