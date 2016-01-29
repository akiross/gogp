package hc

import "fmt"

type Solution interface {
	BetterThan(Solution) bool
	Mutate()
	Fitness() float64
	String() string
	Copy() Solution
}

type Configuration interface {
	RandomSolution() Solution
	NeighborhoodSize() int
}

// Take a solution and perform one step of HC
func HillStep(N int, sol Solution) (Solution, bool) {
	changes := false
	// Pick best among the N neighbors
	for i := 0; i < N; i++ {
		// Generate neighbor
		nbor := sol.Copy()
		nbor.Mutate()
		// Update if better
		if nbor.BetterThan(sol) {
			sol = nbor
			changes = true
		}
	}
	return sol, changes
}

func HillClimbingStart(start Solution, conf Configuration) Solution {
	cs := start
	for i := 0; ; i++ {
		fmt.Println("Iteration", i)
		// Perform one step
		sol, changes := HillStep(conf.NeighborhoodSize(), cs)
		// If no improvement, stop, else move there
		if !changes {
			fmt.Println("There were no changes!")
			return sol
		} else {
			fmt.Println("Yay, found a better one!")
			cs = sol
		}
	}
}

func HillClimbing(conf Configuration) Solution {
	return HillClimbingStart(conf.RandomSolution(), conf)
}
