package sequence

import (
	"github.com/akiross/gogp/util/stats/max"
	"github.com/akiross/gogp/util/stats/min"
	"github.com/akiross/gogp/util/stats/variance"
)

type SequenceStats struct {
	min.Min
	max.Max
	variance.Variance
}

func Create() *SequenceStats {
	return new(SequenceStats)
}

func (ss *SequenceStats) Observe(val float64) {
	ss.Min.Observe(val)
	ss.Max.Observe(val)
	ss.Variance.Accumulate(val)
}

// Clear all the observations
func (ss *SequenceStats) Clear() {
	ss.Min.Clear()
	ss.Max.Clear()
	ss.Variance.Reset()
}
