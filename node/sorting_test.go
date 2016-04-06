package node

import (
	"math/rand"
	"sort"
	"testing"
)

func TestSorting(t *testing.T) {
	n := 1000
	f := make([]float64, n) // Random floats
	c := make([]float64, n) // Copy of f
	d := make([]int, n)     // Indices
	for i := range f {
		f[i] = rand.Float64()
		c[i] = f[i]
		d[i] = i
	}
	t.Log("Unsorted f", f)
	t.Log("Unsorted d", d)
	p := &coupledSlices{f, d}
	sort.Sort(p)
	t.Log("Sorted f", f)
	t.Log("Sorted d", d)
	// Check if it's sorted (descending)
	for i := 1; i < len(f); i++ {
		if f[i] > f[i-1] {
			t.Fail()
		}
	}
	// Check if accessing by index is fine
	for i := 1; i < len(d); i++ {
		if f[i] != c[d[i]] {
			t.Fail()
		}
	}

}
