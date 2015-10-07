package ga

import (
	"sync"
)

// This is a version of Select that is a stage in a pipeline. Will provide pointers to NEW individuals
func GenSelect(pop Population, selectionSize int) <-chan Individual {
	// A channel for output individuals
	out := make(chan Individual, selectionSize)
	go func() {
		sel, _ := pop.Select(selectionSize)
		for i := range sel {
			out <- sel[i]
		}
		/*
			for i := 0; i < selectionSize; i++ {
				// Pick an initial (pointer to) random individual
				best := pop.Get(rand.Intn(pop.Size()))
				// Select other players and select the best
				for j := 1; j < tournSize; j++ {
					maybe := pop.Get(rand.Intn(pop.Size()))
					if pop.IndividualCompare(maybe, best) {
						best = maybe
					}
				}
				sel := best.Copy() //&Individual{best.node.Copy(), best.fitness, best.fitIsValid}
				out <- sel
			}*/
		close(out)
	}()
	return out
}

func GenCrossover(in <-chan Individual, pCross float64) <-chan Individual {
	out := make(chan Individual)
	go func() {
		// Continue forever
		for {
			// Take one item and if we got a real item
			i1, ok := <-in
			if ok {
				i2, ok := <-in
				if ok {
					// We got two items! Crossover
					i1.Crossover(pCross, i2)
					out <- i1
					out <- i2
				} else {
					// Can't crossover a single item
					out <- i1
				}
			} else {
				close(out)
				break
			}
		}
	}()
	return out
}

func GenMutate(in <-chan Individual, pMut float64) <-chan Individual {
	out := make(chan Individual)
	go func() {
		for ind := range in {
			ind.Mutate(pMut)
			out <- ind
		}
		close(out)
	}()
	return out
}

func FanIn(in ...<-chan Individual) <-chan Individual {
	out := make(chan Individual)
	// Used to wait that a set of goroutines finish
	var wg sync.WaitGroup
	// Function that copy from one channel to out
	emitter := func(c <-chan Individual) {
		for i := range c {
			out <- i
		}
		wg.Done() // Signal the group
	}
	// How many goroutines to wait for
	wg.Add(len(in))
	// Start the routines
	for _, c := range in {
		go emitter(c)
	}

	// Wait for the emitters to finish, then close
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// Collects all the individuals from a channel and return a slice. Size is a hint for performances
func Collector(in <-chan Individual, size int) []Individual {
	pop := make([]Individual, 0, size)
	for ind := range in {
		pop = append(pop, ind)
	}
	return pop
}
