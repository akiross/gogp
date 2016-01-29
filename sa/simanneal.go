package sa

import (
	"fmt"
	"math/rand"
)

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
	MaxMoves() int
}

func CoinAnnealing(conf Configuration) Solution {
	climber := conf.RandomSolution()
	best := climber
	for i, moves := 0, 0; moves < conf.MaxMoves(); i++ {
		if i%50 == 0 {
			fmt.Println("Generation", i, "doing move", moves)
		}

		nbor := climber.Copy()
		nbor.Mutate()
		// Check for improvements
		if nbor.BetterThan(climber) || rand.Intn(2) == 0 {
			climber = nbor
			moves++
		}
		if nbor.BetterThan(best) {
			fmt.Println("Found best ever at generation", i, "at move", moves)
			best = nbor
			// Reset moves count
			moves = 0
		}
	}
	return best
}
