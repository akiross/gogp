package main

import (
	//	"ale-re.net/phd/gogp"
	"flag"
	"fmt"
	"math/rand"
	"time"
)

type Individual struct {
	genotype   [10]bool
	fitness    float32
	fitIsValid bool
}

func (i *Individual) String() string {
	s := "["
	for _, g := range i.genotype {
		if g {
			s += "1 "
		} else {
			s += "0 "
		}
	}
	return fmt.Sprint(s[0:len(s)-1]+"] = ", i.fitness)
}

// Minimization problem FIXME this may change
func BetterThan(i, j *Individual) bool {
	return i.fitness < j.fitness
}

// Build a random individual
func CreateRandomIndividual() *Individual {
	var ind Individual
	ind.fitness = 0
	ind.fitIsValid = false
	for i := 0; i < len(ind.genotype); i++ {
		ind.genotype[i] = rand.Intn(2) == 0
	}
	return &ind
}

// Crossover two individuals
func Crossover(a *Individual, b *Individual) (*Individual, *Individual) {
	// Find crossover point (single)
	p := rand.Intn(len(a.genotype)-1) + 1
	// Define new individuals
	var ab, ba Individual
	// Copy data
	for i := 0; i < p; i++ {
		ab.genotype[i], ba.genotype[i] = a.genotype[i], b.genotype[i]
	}
	for i := p; i < len(a.genotype); i++ {
		ab.genotype[i], ba.genotype[i] = b.genotype[i], a.genotype[i]
	}
	// Fitness is unknown and invalid, default
	return &ab, &ba
}

func Mutation(a *Individual, pMut float64) *Individual {
	var b Individual
	for i := 0; i < len(a.genotype); i++ {
		if rand.Float64() < pMut {
			b.genotype[i] = !a.genotype[i]
		} else {
			b.genotype[i] = a.genotype[i]
		}
	}
	return &b
}

type ParamError struct {
	what string
}

func (e *ParamError) Error() string {
	return e.what
}

// Select individuals from population
func TournamentSelection(pop []*Individual, selectionSize, tournSize int) ([]*Individual, error) {
	if (selectionSize < 1) || (tournSize < 1) {
		return nil, &ParamError{"Cannot have selectionSize < 1 or tournSize < 1"}
	}
	// Slice to store the new population
	newPop := make([]*Individual, selectionSize)
	// Perform tournaments
	for i := 0; i < selectionSize; i++ {
		// Pick an initial best
		best := pop[rand.Intn(len(pop))]
		// Select other players
		for j := 1; j < tournSize; j++ {
			maybe := pop[rand.Intn(len(pop))]
			if BetterThan(maybe, best) {
				best = maybe
			}
		}
		// Save winner
		newPop[i] = best
	}
	return newPop, nil
}

func CreatePopulation(popSize int) []*Individual {
	pop := make([]*Individual, popSize)
	for i := range pop {
		pop[i] = CreateRandomIndividual()
	}
	return pop
}

func Fitness(i *Individual) float32 {
	var f float32
	for _, g := range i.genotype {
		if g {
			f += 1
		}
	}
	return f
}

func main() {
	seed := flag.Int64("seed", time.Now().UTC().UnixNano(), "Seed for RNG")
	numGen := flag.Int("gen", 10, "Number of generations")
	popSize := flag.Int("pop", 10, "Size of population")
	tournSize := flag.Int("tourn", 3, "Tournament size")
	pCross := flag.Float64("pCross", 0.9, "Crossover probability")
	pMut := flag.Float64("pMut", 0.1, "Bit mutation probability")

	flag.Parse()

	//	fmt.Println(gogp.Reverse("ciao"))

	rand.Seed(*seed)

	// Create population
	pop := CreatePopulation(*popSize)

	// Save the best individual found so far
	var best *Individual = pop[0]

	// Loop until max number of generations is reached
	for g := 0; g < *numGen; g++ {
		fmt.Print("Generation ", g)

		// Compute fitness for every individual with no fitness
		fitnessEval := 0 // FIXME would be better if this was not always == *popSize
		for _, indiv := range pop {
			if !indiv.fitIsValid {
				indiv.fitness = Fitness(indiv)
				indiv.fitIsValid = true
				fitnessEval++
			}
			if BetterThan(indiv, best) {
				best = indiv
			}
		}
		fmt.Println(" fit evals", fitnessEval)

		// Apply selection
		sel, _ := TournamentSelection(pop, len(pop), *tournSize)

		// Crossover and mutation
		for i := 0; i < len(sel)-1; i += 2 {
			if rand.Float64() < *pCross {
				sel[i], sel[i+1] = Crossover(sel[i], sel[i+1])
			}
			sel[i] = Mutation(sel[i], *pMut)
			sel[i+1] = Mutation(sel[i+1], *pMut)
		}

		// Replace old population
		pop = sel
	}

	fmt.Println("Best individual", best)
}
