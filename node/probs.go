package node

import (
	"github.com/gonum/floats"
	"math/rand"
	"sort"
)

// Normalize the vector of value, summing them and dividing each by the total
func normalSlice(v []float64) {
	tot := floats.Sum(v)
	floats.Scale(1.0/tot, v)
}

// Compute CDF in probs, sorting it and inds as well
// Returns a slice with sorted probabilities
func computeCDFinPlace(probs []float64, inds []int) []float64 {
	// Sort and cumulate the probabilities and indices
	sorted := &coupledSlices{probs, inds}
	sort.Sort(sorted)
	// Copy the sorted probabilities
	cp := make([]float64, len(probs))
	copy(cp, probs)
	for i := 1; i < len(probs); i++ {
		probs[i] += probs[i-1]
	}
	return cp
}

// Extract item and returns its position
func extractCFDinPlace(cdf []float64) int {
	// Pick a random element
	p := rand.Float64()
	var i int
	for i = 0; i < len(cdf); i++ {
		if p < cdf[i] {
			return i
		}
	}
	return i
}
