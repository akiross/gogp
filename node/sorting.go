package node

type coupledSlices struct {
	Target []float64
	Side   []int
}

func (c *coupledSlices) Len() int {
	return len(c.Target)
}

func (c *coupledSlices) Less(i, j int) bool {
	return c.Target[i] > c.Target[j]
}

func (c *coupledSlices) Swap(i, j int) {
	tt := c.Target[i]
	c.Target[i] = c.Target[j]
	c.Target[j] = tt

	st := c.Side[i]
	c.Side[i] = c.Side[j]
	c.Side[j] = st
}
