package stats

import (
	"encoding/json"
	"fmt"
	"github.com/akiross/gogp/apps/base"
	"github.com/akiross/gogp/ga"
	"github.com/akiross/gogp/node"
	"github.com/akiross/gogp/util/stats/counter"
	"github.com/akiross/gogp/util/stats/max"
	"github.com/akiross/gogp/util/stats/min"
	"github.com/akiross/gogp/util/stats/variance"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

type Stats struct {
	basedir, basename    string
	snapCount            int
	obsCount             int // Number of observation
	depth, size, fitness variance.Variance
	min                  min.Min             // Min fitness
	max                  max.Max             // Max fitness
	xoImpr, mutImpr      counter.BoolCounter // Count how often xo and mut improve
	lastTime             time.Time           // Time of last snapshot
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
		depths[i] = node.Depth(pop.Pop[i].Node)
	}
	return depths
}

// Keep track of changes
func (stats *Stats) Observe(pop *base.Population) {
	stats.obsCount += 1
	for i := range pop.Pop {
		// Accumulate depth
		depth := node.Depth(pop.Pop[i].Node)
		stats.depth.Accumulate(float64(depth))
		// Accumulate number of nodes
		size := node.Size(pop.Pop[i].Node)
		stats.size.Accumulate(float64(size))
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

func writeIndividual(ind ga.Individual, outFile string) {
	f, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	bestJson, jsonErr := json.Marshal(ind)
	if jsonErr != nil {
		panic(jsonErr)
	} else {
		f.Write(bestJson)
	}
}

// Another stat: check for correlation between tree depth and tree fitness (deep are good? short are good? what in between?)
// In general, we would like to keep some time-series, but we cannot keep them for every individual or it will take way too much memory!

func (stats *Stats) SaveSnapshot(pop *base.Population, quiet bool, cntKeys, staKeys, intCntKeys []string) (snapName, snapPopName string) {
	timeDelay := time.Since(stats.lastTime)
	stats.lastTime = time.Now()

	// Sort population, to easy reading when printing and drawing
	sort.Sort(pop)
	// Build paths
	prefix := fmt.Sprintf("%v/snapshot/%v-", stats.basedir, stats.basename)
	snapName = fmt.Sprintf(prefix+"snapshot-%v.png", stats.snapCount)
	snapPopName = fmt.Sprintf(prefix+"pop_snapshot-%v.png", stats.snapCount)
	logPrefix := fmt.Sprintf("%v/log/%v-", stats.basedir, stats.basename)
	bestTree := fmt.Sprintf(logPrefix+"tree-%v.json", stats.snapCount)

	writeIndividual(pop.BestIndividual(), bestTree)

	const wideField = 40

	if !quiet {
		if stats.snapCount == 0 {
			fmt.Print("Generation |  Tree depth (mean, stdev) |  Tree size (mean, stdev) |       Fitness (min, mean, max, stdev)       |  XO Improv (abs, rel) | MUT Improv (abs, rel) |    Time delay |")
			for _, k := range staKeys {
				fmt.Printf("%20s", k)
			}
			for _, k := range cntKeys {
				fmt.Printf(" %21s |", k)
			}
			for _, k := range intCntKeys {
				fmt.Printf(" %*s |", wideField, k)
			}
			fmt.Println()
		}
		fmt.Printf("%10v |   %11.4f %11.4f |  %11.4f %11.4f | %10.2f %10.2f %10.2f %10.2f | %10v %10.3f | %10v %10.3f | %13v |",
			stats.obsCount-1,
			stats.depth.PartialMean(), math.Sqrt(stats.depth.PartialVar()),
			stats.size.PartialMean(), math.Sqrt(stats.size.PartialVar()),
			stats.min.Get(), stats.fitness.PartialMean(), stats.max.Get(), math.Sqrt(stats.fitness.PartialVar()),
			stats.xoImpr.AbsoluteFrequency(), stats.xoImpr.RelativeFrequency(),
			stats.mutImpr.AbsoluteFrequency(), stats.mutImpr.RelativeFrequency(),
			fmt.Sprintf("%v", timeDelay),
		)

		for _, k := range staKeys {
			if sst, ok := pop.Set.Statistics[k]; ok {
				fmt.Printf(" %6d %6d %6d %10.6g %10.6g |", sst.Variance.Count(), sst.Min.Get(), sst.Max.Get(), sst.PartialMean(), sst.PartialVarBessel())
				sst.Clear()
			} else {
				fmt.Printf(" %6v %6v %6v %10v %10v |", "-", "-", "-", "-", "-")
			}
		}
		for _, k := range cntKeys {
			if cst, ok := pop.Set.Counters[k]; ok {
				fmt.Printf(" %10d %10.6g |", cst.AbsoluteFrequency(), cst.RelativeFrequency())
				cst.Clear()
			} else {
				fmt.Printf(" %10v %10v |", "-", "-")
			}
		}
		for _, k := range intCntKeys {
			if nst, ok := pop.Set.IntCounters[k]; ok {
				keys := nst.Counted()
				vals := make([]string, len(keys))
				for i, n := range keys {
					vals[i] = fmt.Sprintf("%d:%d", n, nst.AbsoluteFrequency(n))
				}
				fmt.Printf(" %*s |", wideField, strings.Join(vals, ","))
				nst.Clear()
			} else {
				fmt.Printf(" %*s |", wideField, "-")
			}
		}
		fmt.Println()
	}
	// Increment snapshot count
	stats.snapCount++
	return
}
