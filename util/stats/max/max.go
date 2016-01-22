package max

type Max struct {
	max           float64
	count, maxIdx int // how many values visited, which one was the maximum
}

// Clear the observations
func (m *Max) Clear() {
	m.max = 0
	m.count = 0
	m.maxIdx = 0
}

func (m *Max) Observe(v float64) {
	// If we didn't have a maximum, set it
	if m.count == 0 {
		m.maxIdx = 0
		m.max = v
	} else if v > m.max { // Else check if a new max has been found
		m.max = v
		m.maxIdx = m.count
	}
	m.count += 1
}

// Get the maximum value found so far TODO error if none was found
func (m *Max) Get() float64 {
	return m.max
}

// Get the ID (sequence number) of the maximum found
func (m *Max) GetId() int {
	if m.count > 0 {
		return m.maxIdx
	} else {
		return -1
	}
}

// Get the number of visited values
func (m *Max) Count() int {
	return m.count
}
