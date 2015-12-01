package mean

type Mean struct {
	count int     // Number of accumulations
	acc   float64 // Accumulated values
}

func Create() *Mean {
	return new(Mean)
}

// Reset the mean, clearing everything
func (m *Mean) Reset() {
	m.count = 0
	m.acc = 0
}

// Accumulate a value that will be used for the mean
func (m *Mean) Accumulate(v float64) {
	m.acc += v
	m.count += 1
}

// Number of counted accumulation
func (m *Mean) Count() int {
	return m.count
}

// Compute partial mean, using accumulated values until now
func (m *Mean) PartialMean() float64 {
	return m.acc / float64(m.count)
}
