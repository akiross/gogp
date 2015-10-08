package base

import (
	"github.com/akiross/gogp/ga"
	"github.com/akiross/gogp/image/draw2d/imgut"
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
	Set       *Settings
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
	// Save the maxDepth
	origMaxDepth := pop.Set.MaxDepth

	// Divide the pop
	indPerSlice := n / origMaxDepth

	pop.Pop = make([]*Individual, n)
	i := 0

	// Ramped initialization (change max depth)
	for d := 1; d <= pop.Set.MaxDepth; d++ {
		// Set the depth
		pop.Set.MaxDepth = d
		// Initialize the pop
		for j := 0; j < indPerSlice; j++ {
			pop.Pop[i] = new(Individual)
			pop.Pop[i].set = pop.Set
			pop.Pop[i].Initialize()
			i += 1
		}
	}
	for ; i < n; i++ {
		pop.Pop[i] = new(Individual)
		pop.Pop[i].set = pop.Set
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

// Population sorting by fitness
func (pop *Population) Len() int      { return pop.Size() }
func (pop *Population) Swap(i, j int) { pop.Pop[i], pop.Pop[j] = pop.Pop[j], pop.Pop[i] }
func (pop *Population) Less(i, j int) bool {
	fi, fj := pop.Pop[i].fitness, pop.Pop[j].fitness
	return pop.BetterThan(fi, fj)
}

func (pop *Population) Draw(img *imgut.Image, cols, rows int) {
	// From best to worst, draw the images
	for i := range pop.Pop {
		// Draw individual on temporary surface
		pop.Pop[i].Draw(pop.Set.ImgTemp)
		// Copy temporary surface to position
		r, c := i/cols, i%cols
		pop.Set.ImgTemp.Blit(c*pop.Set.ImgTemp.W, r*pop.Set.ImgTemp.H, img)
	}
}
