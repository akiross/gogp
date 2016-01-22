package min

type Min struct {
	min           float64
	count, minIdx int // how many values visited, which one was the minimum
}

// Clear the observations
func (m *Min) Clear() {
	m.min = 0
	m.count = 0
	m.minIdx = 0
}

func (m *Min) Observe(v float64) {
	// If we didn't have a minimum, set it
	if m.count == 0 {
		m.minIdx = 0
		m.min = v
	} else if v < m.min { // else check for a new minima
		m.min = v
		m.minIdx = m.count
	}
	m.count += 1
}

// Get the minimum value found so far TODO error if none was found
func (m *Min) Get() float64 {
	return m.min
}

// Get the ID (sequence number) of the minimum found
func (m *Min) GetId() int {
	if m.count > 0 {
		return m.minIdx
	} else {
		return -1
	}
}

// Get the number of visited values
func (m *Min) Count() int {
	return m.count
}
