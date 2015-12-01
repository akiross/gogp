package mean

import "testing"

func TestMean(t *testing.T) {

	cases := []struct {
		vals     []float64
		expected float64
	}{
		{[]float64{1, 2, 3}, 2},
		{[]float64{600, 470, 170, 430, 300}, 394},
	}

	m := Create()

	for _, c := range cases {
		for _, v := range c.vals {
			m.Accumulate(v)
		}
		if m.PartialMean() != c.expected {
			t.Errorf("Expected mean %v, got %v", c.expected, m.PartialMean())
		}
		m.Reset()

	}
}
