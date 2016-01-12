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
	i := 0 // Initialized individuals

	// Ramped initialization (change max depth)
	for d := 1; d <= origMaxDepth; d++ {
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
	// Add the missing in the last slice
	for ; i < n; i++ {
		pop.Pop[i] = new(Individual)
		pop.Pop[i].set = pop.Set
		pop.Pop[i].Initialize()
	}
}

func (pop *Population) Select(n int, gen float32) ([]ga.Individual, error) {
	selectionSize, tournSize := n, pop.TournSize
	divSize := 7
	randomCount := int(float32(selectionSize) * 0.5 * (1 - gen))

	if (selectionSize < 1) || (tournSize < 1) {
		return nil, &ParamError{"Cannot have selectionSize < 1 or tournSize < 1"}
	}

	// Each individual is selected randomly
	randomSampler := func() *Individual {
		return pop.Pop[rand.Intn(len(pop.Pop))]
	}

	// Each individual is selected via a tournament for "maximum diversity"
	subMaxDiversitySampler := func(subPopSize int) *Individual {
		// Get images of individuals
		subPop := make([]*Individual, subPopSize)
		subPopImages := make([]*imgut.Image, subPopSize)
		for i := range subPop {
			subPop[i] = randomSampler()
			subPopImages[i] = subPop[i].ImgTemp
		}

		// Compute the average image
		avgImage := imgut.Average(subPopImages)

		// Once the average is computed, pick a random individual
		best := 0
		// And compute its distance from the average
		bestDist := imgut.PixelRMSE(subPop[best].ImgTemp, avgImage)
		// Select other players and get the best
		for i := 1; i < subPopSize; i++ {
			// Compute distance from average
			maybeDist := imgut.PixelRMSE(subPop[i].ImgTemp, avgImage)
			// If distance increases, we are maximizing diversity
			if maybeDist > bestDist {
				bestDist = maybeDist
				best = i
			}
		}
		// Return the most different individual from the average image
		return subPop[best]
	}

	// This function select the sample strategy depending on the current i
	individualSampler := func(i int) *Individual {
		if i < randomCount {
			return randomSampler()
		} else {
			return subMaxDiversitySampler(divSize)
		}
	}

	// Slice to store the new population
	newPop := make([]ga.Individual, selectionSize)
	for i := 0; i < selectionSize; i++ {
		// Pick an initial (pointer to) random individual
		best := individualSampler(i)
		// Select other players and select the best
		for j := 1; j < tournSize; j++ {
			maybe := individualSampler(i)
			if pop.BetterThan(maybe.fitness, best.fitness) {
				best = maybe
			}
		}
		newPop[i] = best.Copy()
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
		imgPtr := pop.Pop[i].ImgTemp //pop.Set.ImgTemp)
		// Draw individual on temporary surface
		pop.Pop[i].Draw(imgPtr)
		// Copy temporary surface to position
		r, c := i/cols, i%cols
		imgPtr.Blit(c*imgPtr.W, r*imgPtr.H, img)
		//		pop.Set.ImgTemp.Blit(c*pop.Set.ImgTemp.W, r*pop.Set.ImgTemp.H, img)
	}
}
