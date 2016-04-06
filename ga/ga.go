package ga

import "fmt"

// A Fitness is a real measure
type Fitness float64

// Individuals must satisfy this interface
type Individual interface {
	Copy() Individual              // Copy the individual
	Crossover(float64, Individual) // Crossover with given probability
	Evaluate() Fitness             // Evaluate and return fitness, possibly caching
	Invalidate()                   // Force invalidation of fitness
	FitnessValid() bool            // True if fitness is valid
	Fitness() Fitness              // Return fitness of individual, eventually if necessary
	Initialize()                   // Be initializable
	Mutate(p float64)              // Mutate with given probability
	fmt.Stringer                   // Be convertible to string
	//	SetMetadata(key, value string) // Set metadata for this individual
}

type Population interface {
	Evaluate() int                                   // Evaluate fitnesses, return evaluated ones
	Get(i int) Individual                            // Get a pointer to ith individual
	Initialize(n int)                                // Build N individuals
	Size() int                                       // Get the number of individuals
	Select(n int, gen float32) ([]Individual, error) // Select N individuals at given percentage of evolution process
	BestIndividual() Individual                      // Return the best individual in current population
}

type FitnessComparator interface {
	BetterThan(a, b Fitness) bool
	IndividualCompare(a, b Individual) bool
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
