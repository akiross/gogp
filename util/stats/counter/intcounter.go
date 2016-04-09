package counter

import "sort"

type IntCounter struct {
	intCount map[int]int
	totCount int
}

func (c *IntCounter) Clear() {
	c.intCount = make(map[int]int)
	c.totCount = 0
}

func (c *IntCounter) Count(v int) {
	if c.intCount == nil {
		c.intCount = make(map[int]int)
	}
	c.intCount[v]++
	c.totCount++
}

// Returns the value that have been counted (not their frequencies)
func (c *IntCounter) Counted() []int {
	keys := make(sort.IntSlice, len(c.intCount))
	i := 0
	for k := range c.intCount {
		keys[i] = k
		i++
	}
	keys.Sort()
	return keys
}

func (c *IntCounter) TotalCounts() int {
	return c.totCount
}

func (c *IntCounter) AbsoluteFrequency(v int) int {
	return c.intCount[v]
}

func (c *IntCounter) RelativeFrequency(v int) float64 {
	return float64(c.intCount[v]) / float64(c.totCount)
}
