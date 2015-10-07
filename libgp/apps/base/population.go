package base

import (
	"github.com/akiross/libgp/ga"
	"math/rand"
)

type ParamError struct {
	what string
}

func (e *ParamError) Error() string {
	return e.what
}

type Population struct {
	best      *Individual
	Pop       []*Individual
	ops       *Settings
	TournSize int
	ga.MinProblem
}

func (pop *Population) BestIndividual() ga.Individual {
	return pop.best
}

func (pop *Population) Get(i int) ga.Individual {
	return pop.Pop[i]
}
func (pop *Population) Size() int {
	return len(pop.Pop)
}

// Evaluate population fitness, return number of evaluations
func (pop *Population) Evaluate() (fitnessEval int) {
	for i := range pop.Pop {
		var ind *Individual = pop.Pop[i]
		if !ind.FitnessValid() {
			ind.Evaluate()
			fitnessEval++
		}
		if pop.best == nil || pop.BetterThan(ind.fitness, pop.best.fitness) {
			pop.best = ind
		}
	}
	return
}

func (pop *Population) Initialize(n int) {
	pop.Pop = make([]*Individual, n)
	for i := range pop.Pop {
		pop.Pop[i] = new(Individual)
		pop.Pop[i].Initialize()
	}
}

func (pop *Population) Select(n int) ([]ga.Individual, error) {
	selectionSize, tournSize := n, pop.TournSize
	if (selectionSize < 1) || (tournSize < 1) {
		return nil, &ParamError{"Cannot have selectionSize < 1 or tournSize < 1"}
	}
	// Slice to store the new population
	newPop := make([]ga.Individual, selectionSize)
	for i := 0; i < selectionSize; i++ {
		// Pick an initial (pointer to) random individual
		best := pop.Pop[rand.Intn(len(pop.Pop))]
		// Select other players and select the best
		for j := 1; j < tournSize; j++ {
			maybe := pop.Pop[rand.Intn(len(pop.Pop))]
			if pop.BetterThan(maybe.fitness, best.fitness) {
				best = maybe
			}
		}
		newPop[i] = &Individual{best.Node.Copy(), best.fitness, best.fitIsValid, best.set}
	}
	return newPop, nil
}
