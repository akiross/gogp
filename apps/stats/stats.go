package stats

import (
	"fmt"
	"github.com/akiross/gogp/apps/base"
	"github.com/akiross/gogp/ga"
	"github.com/akiross/gogp/util/stats/counter"
	"github.com/akiross/gogp/util/stats/max"
	"github.com/akiross/gogp/util/stats/min"
	"github.com/akiross/gogp/util/stats/variance"
	"math"
	"sort"
	"time"
)

type Stats struct {
	basedir, basename string
	snapCount         int
	obsCount          int // Number of observation
	depth, fitness    variance.Variance
	min               min.Min         // Min fitness
	max               max.Max         // Max fitness
	xoImpr, mutImpr   counter.Counter // Count how often xo and mut improve
	lastTime          time.Time       // Time of last snapshot
}

func Create(basedir, basename string) *Stats {
	var stats Stats
	stats.basedir = basedir
	stats.basename = basename
	stats.lastTime = time.Now()
	return &stats
}

// Returns a slice with depths for every individual in the population
func (stats *Stats) PopulationDepths(pop *base.Population) []int {
	depths := make([]int, len(pop.Pop))
	for i := range pop.Pop {
		depths[i] = pop.Pop[i].Node.Depth()
	}
	return depths
}

// Keep track of changes
func (stats *Stats) Observe(pop *base.Population) {
	stats.obsCount += 1
	for i := range pop.Pop {
		// Accumulate depth
		depth := pop.Pop[i].Node.Depth()
		stats.depth.Accumulate(float64(depth))

		// Accumulate fitness
		fit := float64(pop.Pop[i].Fitness())
		stats.fitness.Accumulate(fit)
		// Store min and max fitnesses
		stats.min.Observe(fit)
		stats.max.Observe(fit)
	}
}

// Count how many time new fitness is better than old fitness
func (stats *Stats) ObserveCrossoverFitness(newFit, oldFit ga.Fitness) {
	stats.xoImpr.Count(newFit < oldFit) // FIXME this one should use ga.FitnessComparator
}

func (stats *Stats) ObserveMutationFitness(newFit, oldFit ga.Fitness) {
	stats.mutImpr.Count(newFit < oldFit)
}

// Another stat: check for correlation between tree depth and tree fitness (deep are good? short are good? what in between?)
// In general, we would like to keep some time-series, but we cannot keep them for every individual or it will take way too much memory!

func (stats *Stats) SaveSnapshot(pop *base.Population, quiet bool) (snapName, snapPopName string) {
	timeDelay := time.Since(stats.lastTime)
	stats.lastTime = time.Now()

	// Sort population, to easy reading when printing and drawing
	sort.Sort(pop)
	// Build paths
	prefix := fmt.Sprintf("%v/snapshot/%v-", stats.basedir, stats.basename)
	snapName = fmt.Sprintf(prefix+"snapshot-%v.png", stats.snapCount)
	snapPopName = fmt.Sprintf(prefix+"pop_snapshot-%v.png", stats.snapCount)

	if !quiet {
		// TODO XXX FIXME this will be re-introduced when individuals support informative pretty-printing
		fmt.Println("Saving best individual snapshot", snapName)
		fmt.Println(pop.BestIndividual())

		///////////////                                                                                                   111111
		///////////////         111111111122222222223333333333444444444455555555556666666666777777777788888888889999999999000000
		///////////////123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345
		if stats.snapCount == 0 {
			fmt.Println("Generation |  Tree depth (mean, stdev) |       Fitness (min, mean, max, stdev)       |  XO Improv (abs, rel) | MUT Improv (abs, rel) |    Time delay |")
		}
		fmt.Printf("%10v |   %11.4f %11.4f | %10.2f %10.2f %10.2f %10.2f | %10v %10.3f | %10v %10.3f | %13v |\n",
			stats.obsCount-1,
			stats.depth.PartialMean(), math.Sqrt(stats.depth.PartialVar()),
			stats.min.Get(), stats.fitness.PartialMean(), stats.max.Get(), math.Sqrt(stats.fitness.PartialVar()),
			stats.xoImpr.AbsoluteFrequency(), stats.xoImpr.RelativeFrequency(),
			stats.mutImpr.AbsoluteFrequency(), stats.mutImpr.RelativeFrequency(),
			fmt.Sprintf("%v", timeDelay),
		)

		//fmt.Println("Statistics:")
		//fmt.Println("  Tree depth (mean, stdev):", stats.depth.PartialMean(), math.Sqrt(stats.depth.PartialVar()))
		//fmt.Println("  Fitness (mean, stdev):", stats.fitness.PartialMean(), math.Sqrt(stats.fitness.PartialVar()))
		//fmt.Println("  Fitness (min, max):", stats.min.Get(), stats.max.Get())
		//fmt.Println("  XO Improv (abs, rel):", stats.xoImpr.AbsoluteFrequency(), stats.xoImpr.RelativeFrequency())
		//fmt.Println("  MUT Improv (abs, rel):", stats.mutImpr.AbsoluteFrequency(), stats.mutImpr.RelativeFrequency())

		// Print the fitnesses for each individual
		//				for kk := range pop.Pop {
		//		fmt.Println(kk, "-th individual has fitness", pop.Pop[kk].Fitness())
		//		}
		//				fmt.Println("Saving pop snapshot", snapPopName)
		//				fmt.Println(pop)
	}

	// Compute advanced statistics
	/*
		if stats.AdvancedStats {
			// Compute average depth
			fmt.Print("ADVSTAT: Depths:")
			totDepth := 0
			for i := range pop.Pop {
				dep := pop.Pop[i].Node.Depth()
				fmt.Print(", ", dep)
				totDepth += dep
			}
			fmt.Printf("\nAverage depth: %f\n", float64(totDepth)/float64(len(pop.Pop)))
		}
	*/

	/*
		// Save best individual
		pop.BestIndividual().(*base.Individual).Draw(settings.ImgTemp)
		settings.ImgTemp.WritePNG(snapName)
		// Save pop images
		pop.Draw(imgTempPop, pImgCols, pImgRows)
		imgTempPop.WritePNG(snapPopName)
	*/
	// Increment snapshot count
	stats.snapCount++
	return
}
