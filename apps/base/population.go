package base

import (
	"github.com/akiross/gogp/ga"
	"github.com/akiross/gogp/image/draw2d/imgut"
)

type ParamError struct {
	what string
}

func (e *ParamError) Error() string {
	return e.what
}

type Population struct {
	best *Individual
	Pop  []*Individual
	Set  *Settings
	//TournSize int
	//ga.MinProblem
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
		if pop.best == nil || pop.Set.BetterThan(ind.fitness, pop.best.fitness) {
			pop.best = ind
		}
	}
	return
}

func (pop *Population) Initialize(n int) {
	pop.Pop = make([]*Individual, n) // Build population
	i := 0                           // Initialized individuals

	if pop.Set.Ramped { // Ramped init
		// Save the maxDepth
		origMaxDepth := pop.Set.MaxDepth
		// Use ramped half and half if possible
		if origMaxDepth > 0 {
			// Divide the pop
			indPerSlice := n / origMaxDepth

			// Ramped initialization (warning: changes max depth) FIXME this is a side effect, not really nice...
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
		}
	}
	// Add the missing individuals in the last slice
	for ; i < n; i++ {
		pop.Pop[i] = new(Individual)
		pop.Pop[i].set = pop.Set
		pop.Pop[i].Initialize()
	}
}

func MakeSelectTourn(tournSize int, betterFit func(a, b ga.Fitness) bool) func([]*Individual, int) []ga.Individual {
	return func(oldPop []*Individual, selectionSize int) []ga.Individual {
		// Slice to store the new population
		newPop := make([]ga.Individual, selectionSize)

		for i := 0; i < selectionSize; i++ {
			// Sample random individuals for tournament
			players := SampleRandom(oldPop, tournSize)
			// Perform tournament using fitness
			best := SampleTournament(players, betterFit)
			// Save best to population
			newPop[i] = best.Copy()
		}

		return newPop
	}
}

func MakeSelectRMAD(tournSize, divSize int, betterFit func(a, b ga.Fitness) bool) func([]*Individual, int) []ga.Individual {
	return func(oldPop []*Individual, selectionSize int) []ga.Individual {
		// Slice to store the new population
		newPop := make([]ga.Individual, selectionSize)
		for i := 0; i < selectionSize; i++ {
			// Sample some individuals using diversity tournament
			sample := make([]*Individual, tournSize)
			for i := range sample {
				// Sample some random individuals for tournament
				players := SampleRandom(oldPop, divSize)
				// Do fitness tournament and save winner to sample
				sample[i] = SampleSMDTournament(players)
			}
			// Now perform tournament using fitness
			best := SampleTournament(sample, betterFit)
			// Save best to population
			newPop[i] = best.Copy()
		}
		return newPop
	}
}

func MakeSelectIRMAD(tournSize, divSize int, betterFit func(a, b ga.Fitness) bool) func([]*Individual, int) []ga.Individual {
	return func(oldPop []*Individual, selectionSize int) []ga.Individual {
		// Slice to store the new population
		newPop := make([]ga.Individual, selectionSize)
		for i := 0; i < selectionSize; i++ {
			// Sample some individuals using fitness tournament
			sample := make([]*Individual, divSize)
			for i := range sample {
				// Sample some random individuals for tournament
				players := SampleRandom(oldPop, tournSize)
				// Do fitness tournament and save winner to sample
				sample[i] = SampleTournament(players, betterFit)
			}
			// Now perform tournament using diversity
			best := SampleSMDTournament(sample)
			// Save best to population
			newPop[i] = best.Copy()
		}
		return newPop
	}
}

func (pop *Population) Select(n int, gen float32) ([]ga.Individual, error) {
	if n < 1 {
		return nil, &ParamError{"Cannot have selectionSize < 1"}
	}

	return pop.Set.Select(pop.Pop, n), nil

	/*
		// Slice to store the new population
		newPop := make([]ga.Individual, selectionSize)

		switch "Tournament" {
		case "Tournament":
			for i := 0; i < selectionSize; i++ {
				// Sample random individuals for tournament
				players := SampleRandom(pop.Pop, tournSize)
				// Perform tournament using fitness
				best := SampleTournament(players, pop.Set.BetterThan)
				// Save best to population
				newPop[i] = best.Copy()
			}
		case "RMAD": // Relative Maximum Average Diversity
			for i := 0; i < selectionSize; i++ {
				// Sample some individuals using diversity tournament
				sample := make([]*Individual, tournSize)
				for i := range sample {
					// Sample some random individuals for tournament
					players := SampleRandom(pop.Pop, divSize)
					// Do fitness tournament and save winner to sample
					sample[i] = SampleSMDTournament(players)
				}
				// Now perform tournament using fitness
				best := SampleTournament(sample, pop.Set.BetterThan)
				// Save best to population
				newPop[i] = best.Copy()
			}
		case "IRMAD": // Inverse RMAD
			for i := 0; i < selectionSize; i++ {
				// Sample some individuals using fitness tournament
				sample := make([]*Individual, divSize)
				for i := range sample {
					// Sample some random individuals for tournament
					players := SampleRandom(pop.Pop, tournSize)
					// Do fitness tournament and save winner to sample
					sample[i] = SampleTournament(players, pop.Set.BetterThan)
				}
				// Now perform tournament using diversity
				best := SampleSMDTournament(sample)
				// Save best to population
				newPop[i] = best.Copy()
			}
		default: // Error
			panic("Invalid selection method!")
		}

		return newPop, nil
	*/
}

// Population sorting by fitness
func (pop *Population) Len() int      { return pop.Size() }
func (pop *Population) Swap(i, j int) { pop.Pop[i], pop.Pop[j] = pop.Pop[j], pop.Pop[i] }
func (pop *Population) Less(i, j int) bool {
	fi, fj := pop.Pop[i].fitness, pop.Pop[j].fitness
	return pop.Set.BetterThan(fi, fj)
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
