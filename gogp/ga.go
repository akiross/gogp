package gogp

import "fmt"

// A Fitness is a real measure
type Fitness float64

// Compare two fitnesses
type FitnessComparator interface {
	BetterThan(a, b Fitness) bool
}

// Individuals must satisfy this interface
type Individual interface {
	Crossover(float64, *Individual) // Crossover with given probability
	Evaluate()                      // Evaluate and cache fitness
	FitnessValid() bool             // True if fitness is valid
	Initialize()                    // Be initializable
	Mutate(p float64)               // Mutate with given probability
	fmt.Stringer                    // Be convertible to string
}

type Population interface {
	Evaluate() int                      // Evaluate fitnesses, return evaluated ones
	Initialize(n int)                   // Build N individuals
	Select(n int) ([]Individual, error) // Select N individuals
	BestIndividual() *Individual        // Return the best individual in current population
}

type MinProblem struct {
	FitnessComparator
}

type MaxProblem struct {
	FitnessComparator
}

func (p *MinProblem) BetterThan(a, b Fitness) bool {
	return a < b
}

func (p *MaxProblem) BetterThan(a, b Fitness) bool {
	return b > a
}
