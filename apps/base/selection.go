package base

/* This package contains selection operators to be used with the population */

import (
	"github.com/akiross/gogp/ga"
	"github.com/akiross/gogp/image/draw2d/imgut"
	"math/rand"
)

// Randomy samples n individuals from the specified population
func SampleRandom(pop []*Individual, n int) []*Individual {
	samp := make([]*Individual, n)
	for i := range samp {
		samp[i] = pop[rand.Intn(len(pop))]
	}
	return samp
}

// Given the individuals, perform a single tournament selection picking the
// best individual among them
func SampleTournament(sample []*Individual, betterFit func(a, b ga.Fitness) bool) *Individual {
	// Best individual found so far
	b := 0
	// Search the best individual among them
	for i := range sample {
		if betterFit(sample[i].Fitness(), sample[b].Fitness()) {
			b = i // Save best individual
		}
	}
	return sample[b].Copy().(*Individual)
}

// Given the individuals, perform a single "Sub-Max-Diversity" tournament of
// given size. The selected individual is the least different from the average
// of the provided population.
func SampleSMDTournament(sample []*Individual) *Individual {
	// Don't perform the full procedure if only one individual
	if len(sample) == 1 {
		return sample[0].Copy().(*Individual)
	}

	// Get images of sample
	simgs := make([]*imgut.Image, len(sample))
	for i := range sample {
		// TODO make sure this doesn't require a re-evaluation
		simgs[i] = sample[i].ImgTemp
	}

	// Compute the average image
	avgImage := imgut.Average(simgs)

	// Once the average is computed, pick a random individual
	b := 0
	// And compute its distance from the average
	bdist := imgut.PixelRMSE(sample[b].ImgTemp, avgImage)
	// Select other players and get the best
	for i := range sample {
		// Compute distance from average
		dist := imgut.PixelRMSE(sample[i].ImgTemp, avgImage)
		// If distance increases, we are maximizing diversity
		if dist > bdist {
			bdist = dist
			b = i
		}
	}

	return sample[b].Copy().(*Individual)
}

/*
func SelectTournament(pop *Population, int n, gen float32) ([]ga.Individual, error) {
}

func SelectSMDTournament(pop *Population, int n, gen float32) ([]ga.Individual, error) {
}

func (pop *Population) Select(n int, gen float32) ([]ga.Individual, error) {
	selectionSize, tournSize := n, pop.TournSize
	divSize := 7
	randomCount := int(float32(selectionSize) * 0.5 * (1 - gen))

	if (selectionSize < 1) || (tournSize < 1) {
		return nil, &ParamError{"Cannot have selectionSize < 1 or tournSize < 1"}
	}

	// Each individual is selected via a tournament for "maximum diversity"
	subMaxDiversitySampler := func(subPopSize int) *Individual {
		// Get images of individuals
		subPop := make([]*Individual, subPopSize)
		subPopImages := make([]*imgut.Image, subPopSize)
		for i := range subPop {
			subPop[i] = sampleRandom(pop.Pop)
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
*/
