package counter

// Counter keeps track of the frequency of binary events
type Counter struct {
	trueCount, totCount int
}

func (c *Counter) Clear() {
	c.trueCount = 0
	c.totCount = 0
}

func (c *Counter) Count(v bool) {
	if v {
		c.trueCount++
	}
	c.totCount++
}

func (c *Counter) AbsoluteFrequency() int {
	return c.trueCount
}

func (c *Counter) RelativeFrequency() float64 {
	return float64(c.trueCount) / float64(c.totCount)
}

// Expected counts the frequencies of binary events agains an expected value
type Expected struct {
	truePos  int // What we want
	trueNeg  int // What we don't want
	falsePos int // What we thought we wanted, but we don't
	falseNeg int // What we thought we don't want, but we do
}

func (c *Expected) Clear() {
	c.truePos = 0
	c.trueNeg = 0
	c.falsePos = 0
	c.falseNeg = 0
}

func (c *Expected) Count(actual, expected bool) {
	if actual {
		if expected {
			c.truePos++
		} else {
			c.falsePos++
		}
	} else {
		if expected {
			c.falseNeg++
		} else {
			c.trueNeg++
		}
	}
}

func (c *Expected) Precision() float64 {
	return float64(c.truePos) / float64(c.truePos+c.falsePos)
}

func (c *Expected) Recall() float64 {
	return float64(c.truePos) / float64(c.truePos+c.trueNeg)
}
