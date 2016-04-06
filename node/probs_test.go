package node

import (
	"math/rand"
	"testing"
)

func TestComputeCFD(t *testing.T) {
	n := 50
	p := make([]float64, n)
	d := make([]int, n)
	for i := range p {
		p[i] = rand.Float64()
		d[i] = i
	}
	// Normalize and compute CDF
	normalSlice(p)
	computeCDFinPlace(p, d)
	if p[len(p)-1] < 1 {
		t.Error("Failing compute CDF", p)
	}
}

func TestExtractCDF(t *testing.T) {
	n := 10
	f := make([]float64, n)
	d := make([]int, n)
	for i := range f {
		f[i] = rand.Float64()
		d[i] = i
	}
	normalSlice(f)
	p := computeCDFinPlace(f, d)
	l := 10000000
	m := make([]float64, n) // Counts
	for i := 0; i < l; i++ {
		k := extractCFDinPlace(f)
		m[k] += 1.0 / float64(l)
	}
	// Compare to 2nd decimal digit
	t.Log(m)
	t.Log(p)
	for i := range m {
		x, y := int(m[i]*100), int(p[i]*100)
		if x != y {
			t.Fail()
		}
	}
}
