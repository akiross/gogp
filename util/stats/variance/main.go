package variance

import "github.com/akiross/gogp/util/stats/mean"

type Variance struct {
	mean.Mean         // Embedded mean
	sqAcc     float64 // Squared accumulator
}

func Create() *Variance {
	return new(Variance)
}

// Reset clearning everything accumulated
func (v *Variance) Reset() {
	v.Mean.Reset()
	v.sqAcc = 0
}

// Accumulate value to compute mean and variance
func (v *Variance) Accumulate(val float64) {
	v.Mean.Accumulate(val)
	v.sqAcc += val * val
}

// Compute variance for the values accumulated until now
func (v *Variance) PartialVar() float64 {
	mean := v.PartialMean()
	return v.sqAcc/float64(v.Count()) - mean*mean
}

// Bessel-corrected version of PartialVar
// To be used when working with samples, and real mean is unknown
func (v *Variance) PartialVarBessel() float64 {
	n := float64(v.Count())
	return v.PartialVar() * n / (n - 1.0)
}
