package sequence

import (
	"github.com/akiross/gogp/util/stats/max"
	"github.com/akiross/gogp/util/stats/min"
	"github.com/akiross/gogp/util/stats/variance"
)

type SequenceStats struct {
	min.Min
	min.Max
}
