package counter

import (
	"math/rand"
	"testing"
)

func TestIntCounter(t *testing.T) {
	var c IntCounter

	c.Count(0)
	if k := c.Counted(); len(k) != 1 {
		t.Error("Wrong number of counted integers")
	} else {
		if c.AbsoluteFrequency(k[0]) != 1 {
			t.Error("Wrong frequency of counted integer")
		}
	}
	c.Count(0)
	c.Count(1)
	if k := c.Counted(); len(k) != 2 {
		t.Error("Wrong number of counted integers")
	} else {
		if c.AbsoluteFrequency(k[0]) != 2 {
			t.Error("Wrong frequency of counted integer 0")
		}
		if c.AbsoluteFrequency(k[1]) != 1 {
			t.Error("Wrong frequency of counted integer 1")
		}
	}

	c.Clear()

	// Test using frequencies
	n, num := make(map[int]int), 100
	for i := 0; i < num; i++ {
		v := rand.Intn(1000)
		n[v]++
		c.Count(v)
	}

	for k := range n {
		if n[k] != c.AbsoluteFrequency(k) {
			t.Error("Different abs frequencies for", k, ":", n[k], c.AbsoluteFrequency(k))
		}
		rf := float64(n[k]) / float64(num)
		if rf != c.RelativeFrequency(k) {
			t.Error("Different rel frequency for", k, ":", rf, c.RelativeFrequency(k))
		}
	}
}

func TestExpected(t *testing.T) {
}
