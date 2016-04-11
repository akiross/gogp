package evolve

import (
	"flag"
	"fmt"
	"github.com/akiross/gogp/apps/base"
	"github.com/akiross/gogp/apps/stats"
	"github.com/akiross/gogp/ga"
	"github.com/akiross/gogp/gp"
	"github.com/akiross/gogp/image/draw2d/imgut"
	"github.com/akiross/gogp/node"
	"github.com/akiross/gogp/util/stats/counter"
	"github.com/akiross/gogp/util/stats/sequence"
	"math"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof" // profiling...
	"time"
)

const (
	tree_init_depth = "tree-init-depth"

	mut_single_event       = "mut-single-event"
	mut_single_improv      = "mut-single-improv"
	mut_single_node_depth  = "mut-single-node-depth"
	mut_single_node_repld  = "mut-single-node-repld"
	mut_single_node_leaves = "mut-single-node-leaves"

	mut_multi_event       = "mut-multi-event"
	mut_multi_improv      = "mut-multi-improv"
	mut_multi_node_depth  = "mut-multi-node-depth"
	mut_multi_node_repld  = "mut-multi-node-repld"
	mut_multi_node_leaves = "mut-multi-node-leaves"

	mut_tree_event       = "mut-subtree-event"
	mut_tree_improv      = "mut-subtree-improv"
	mut_tree_node_depth  = "mut-tree-node-depth"
	mut_tree_node_repld  = "mut-tree-node-repld"
	mut_tree_node_leaves = "mut-tree-node-leaves"

	mut_area_event       = "mut-area-event"
	mut_area_improv      = "mut-area-improv"
	mut_area_node_depth  = "mut-area-node-depth"
	mut_area_node_repld  = "mut-area-node-repld"
	mut_area_node_leaves = "mut-area-node-leaves"

	mut_local_event  = "mut-local-event"
	mut_local_improv = "mut-local-improv"

	mut_count_multi = "mut-count-multiple"
)

func makeMultiMutation(s *base.Settings, multiMut, enbSin, enbNod, enbSub, enbAre, enbLoc bool) func(float64, *base.Individual) bool {
	// La mutazione a che profondità avviene?
	// Quanto è profondo l'albero che vado a generare?
	// Quanto è profondo l'albero che vado a sostituire?
	// Separare foglie e nodi

	countInt := func(name string, v int) {
		if _, ok := s.IntCounters[name]; !ok {
			s.IntCounters[name] = new(counter.IntCounter)
		}
		s.IntCounters[name].Count(v)
	}

	countBool := func(name string, v bool) {
		if _, ok := s.Counters[name]; !ok {
			s.Counters[name] = new(counter.BoolCounter)
		}
		s.Counters[name].Count(v)
	}

	statFuncSin := func(nDepth, replDepth int, isLeaf bool) {
		countInt(mut_single_node_depth, nDepth)
		countInt(mut_single_node_repld, replDepth)
		countBool(mut_single_node_leaves, isLeaf)
	}
	statFuncNod := func(nDepth, replDepth int, isLeaf bool) {
		countInt(mut_multi_node_depth, nDepth)
		countInt(mut_multi_node_repld, replDepth)
		countBool(mut_multi_node_leaves, isLeaf)
	}
	statFuncTree := func(nDepth, replDepth int, isLeaf bool) {
		countInt(mut_tree_node_depth, nDepth)
		countInt(mut_tree_node_repld, replDepth)
		countBool(mut_tree_node_leaves, isLeaf)
	}
	statFuncArea := func(nDepth, replDepth int, isLeaf bool) {
		countInt(mut_area_node_depth, nDepth)
		countInt(mut_area_node_repld, replDepth)
		countBool(mut_area_node_leaves, isLeaf)
	}

	singleMut := node.MakeTreeSingleMutation(s.Functionals, s.Terminals, statFuncSin) // func(*Node)
	nodeMut := node.MakeTreeNodeMutation(s.Functionals, s.Terminals, statFuncNod)     // funcs, terms []gp.Primitive)// func(*Node) int {
	subtrMut := node.MakeSubtreeMutation(s.MaxDepth, s.GenFunc, statFuncTree)
	areaMut := node.MakeSubtreeMutationGuided(s.MaxDepth, s.GenFunc, node.ArityDepthProbComputer, statFuncArea)

	const neighbSize = 5 // Neighborhood size for local search

	return func(pMut float64, ind *base.Individual) bool {
		perm := rand.Perm(5) // Randomly permutate the algorithms to pick
		evCount := 0         // Number of events (mutations performed)
		for _, v := range perm {
			event := rand.Float64() < pMut // Perform mutation?
			switch v {
			case 0:
				if enbSin {
					ind.CountEvent(mut_single_event, event)
					if event {
						evCount++
						fit := ind.Evaluate()
						singleMut(ind.Node)
						newFit := ind.Evaluate()
						ind.CountEvent(mut_single_improv, s.BetterThan(newFit, fit))
					}
					if !multiMut {
						return event
					}
				}
			case 1:
				if enbNod {
					fit := ind.Evaluate()
					event = nodeMut(pMut, ind.Node) != 0
					ind.CountEvent(mut_multi_event, event)
					if event {
						evCount++
						newFit := ind.Evaluate()
						ind.CountEvent(mut_multi_improv, s.BetterThan(newFit, fit))
					}
					if !multiMut {
						return event
					}
				}
			case 2:
				if enbSub {
					ind.CountEvent(mut_tree_event, event)
					if event {
						evCount++
						fit := ind.Evaluate()
						subtrMut(ind.Node)
						newFit := ind.Evaluate()
						ind.CountEvent(mut_tree_improv, s.BetterThan(newFit, fit))
					}
					if !multiMut {
						return event
					}
				}
			case 3:
				if enbAre {
					ind.CountEvent(mut_area_event, event)
					if event {
						evCount++
						fit := ind.Evaluate()
						areaMut(ind.Node)
						newFit := ind.Evaluate()
						ind.CountEvent(mut_area_improv, s.BetterThan(newFit, fit))
					}
					if !multiMut {
						return event
					}
				}
			default:
				if enbLoc {
					// Local search
					ind.CountEvent(mut_local_event, event)
					if event {
						evCount++
						fit := ind.Evaluate()
						for i := 0; i < neighbSize; i++ {
							mutated := ind.Copy().(*base.Individual) // Copy individual
							singleMut(mutated.Node)                  // Apply mutation
							if s.BetterThan(mutated.Fitness(), ind.Fitness()) {
								// If improved, save the individual
								ind.Node = mutated.Node
								// Invalidate!
								ind.Invalidate()
							}
						}
						// Perform singleMut K times
						newFit := ind.Evaluate()
						ind.CountEvent(mut_local_improv, s.BetterThan(newFit, fit))
					}
					if !multiMut {
						return event
					}
				}
			}
			// In the case we don't execute anything, go to the next method
		}
		// If this happens, all the mutations were disabled, or all the mutations
		// were performed when multi mutation was enabled
		if multiMut {
			countInt(mut_count_multi, evCount)
		}
		return evCount > 0
	}
}

func makeCrossover(s *base.Settings) func(float64, *base.Individual, *base.Individual) bool {
	xo := node.MakeTree1pCrossover(s.MaxDepth)
	return func(pCross float64, mate1, mate2 *base.Individual) bool {
		if rand.Float64() < pCross {
			xo(mate1.Node, mate2.Node)
			return true
		} else {
			return false
		}
	}
}

func Evolve(calcMaxDepth func(*imgut.Image) int, fun, ter []gp.Primitive, drawfun func(*base.Individual, *imgut.Image)) {
	startTime := time.Now()

	// Setup options
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	numGen := fs.Int("g", 100, "Number of generations")
	popSize := fs.Int("p", 1000, "Size of population")
	saveInterval := fs.Int("n", 25, "Generations interval between two snapshot saves")
	tournSize := fs.Int("T", 3, "Tournament size")
	pCross := fs.Float64("C", 0.8, "Crossover probability")
	pMut := fs.Float64("M", 0.1, "Bit mutation probability")
	quiet := fs.Bool("q", false, "Quiet mode")
	fElite := fs.Bool("el", false, "Enable elite individual")

	fInitFull := fs.Bool("full", true, "Enable full initialization")
	fInitGrow := fs.Bool("grow", true, "Enable grow initialization")
	fInitRamped := fs.Bool("ramp", true, "Enable ramped initialization")

	fMutSin := fs.Bool("ms", false, "Enable Single Mutation")
	fMutNod := fs.Bool("mn", false, "Enable Node Mutation")
	fMutSub := fs.Bool("mt", false, "Enable Subtree Mutation")
	fMutAre := fs.Bool("ma", false, "Enable Area Mutation")
	fMutLoc := fs.Bool("ml", false, "Enable Local Mutation")

	fMultiMut := fs.Bool("mM", false, "Enable multiple mutations")

	//advStats := fs.Bool("stats", false, "Enable advanced statistics")
	//nps := fs.Bool("nps", false, "Disable population snapshot (no-pop-snap)")
	targetPath := fs.String("t", "", "Target image (PNG) path")
	var basedir, basename string
	cpuProfile := fs.String("cpuprofile", "", "Write CPU profile to file")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "  basedir (string) to be used for saving logs and files\n")
		fmt.Fprintf(os.Stderr, "  basename (string) to be used for saving logs and files\n")
	}

	fs.Parse(os.Args[1:])

	// Check if the argument
	args := fs.Args()

	if len(args) != 2 {
		fs.Usage()
		fmt.Fprintf(os.Stderr, "\nBasename/basedir parameter not specified\n")
		return
	} else {
		basedir = args[0]
		basename = args[1]
	}

	sta := stats.Create(basedir, basename)

	// Build settings
	var settings base.Settings
	// Primitives to use
	settings.Functionals = fun
	settings.Terminals = ter
	// Draw function to use
	settings.Draw = drawfun

	settings.Ramped = *fInitRamped
	fmt.Println("Using ramped initialization")

	// Pick initialization method based on flags
	var genFuncBit func(maxH int, funcs, terms []gp.Primitive) *node.Node

	if *fInitFull && !*fInitGrow {
		fmt.Println("Using init strategy: full")
		genFuncBit = node.MakeTreeFull // Initialize tree using full
	} else if !*fInitFull && *fInitGrow {
		fmt.Println("Using init strategy: balanced grow")
		genFuncBit = node.MakeTreeGrowBalanced // Initialize using grow
	} else {
		fmt.Println("Using init strategy: half-and-half")
		genFuncBit = node.MakeTreeHalfAndHalf // Initialize using both (half and half)
	}
	settings.GenFunc = func(maxDep int) *node.Node {
		t := genFuncBit(maxDep, fun, ter)
		s := settings
		if _, ok := s.IntCounters[tree_init_depth]; !ok {
			s.IntCounters[tree_init_depth] = new(counter.IntCounter)
		}
		s.IntCounters[tree_init_depth].Count(node.Depth(t))
		return t
	}

	// Build statistic map
	settings.Statistics = make(map[string]*sequence.SequenceStats)
	settings.Counters = make(map[string]*counter.BoolCounter)
	settings.IntCounters = make(map[string]*counter.IntCounter)

	// Names of extra statistics
	statsKeys := []string{}
	countersKeys := []string{}
	intCountersKeys := []string{}

	intCountersKeys = append(intCountersKeys, tree_init_depth)

	if *fMutSin {
		countersKeys = append(countersKeys, mut_single_event, mut_single_improv, mut_single_node_leaves)
		intCountersKeys = append(intCountersKeys, mut_single_node_depth, mut_single_node_repld)
	}
	if *fMutNod {
		countersKeys = append(countersKeys, mut_multi_event, mut_multi_improv, mut_multi_node_leaves)
		intCountersKeys = append(intCountersKeys, mut_multi_node_depth, mut_multi_node_repld)
	}
	if *fMutSub {
		countersKeys = append(countersKeys, mut_tree_event, mut_tree_improv, mut_tree_node_leaves)
		intCountersKeys = append(intCountersKeys, mut_tree_node_depth, mut_tree_node_repld)
	}
	if *fMutAre {
		countersKeys = append(countersKeys, mut_area_event, mut_area_improv, mut_area_node_leaves)
		intCountersKeys = append(intCountersKeys, mut_area_node_depth, mut_area_node_repld)
	}
	if *fMutLoc {
		countersKeys = append(countersKeys, mut_local_event, mut_local_improv)
		//		intCountersKeys = append(intCountersKeys, "mut_local_") TODO
	}
	if *fMultiMut {
		intCountersKeys = append(intCountersKeys, mut_count_multi)
	}

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			fmt.Println("ERROR", err)
			panic("Cannot create cpuprofile")
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Load the target
	var err error
	settings.ImgTarget, err = imgut.Load(*targetPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Cannot load image", *targetPath)
		panic("Cannot load image")
	}
	if !*quiet {
		fmt.Println("Image format RGB?", settings.ImgTarget.ColorSpace == imgut.MODE_RGB, settings.ImgTarget.ColorSpace)
	}

	// Compute the right value of maxDepth
	settings.MaxDepth = calcMaxDepth(settings.ImgTarget)
	if !*quiet {
		fmt.Println("For area of", settings.ImgTarget.W*settings.ImgTarget.H, "pixels, max depth is", settings.MaxDepth)
	}

	// Create temporary surface, of same size and mode
	//settings.ImgTemp = imgut.Create(settings.ImgTarget.W, settings.ImgTarget.H, settings.ImgTarget.ColorSpace)
	// Create temporary surface for the entire population
	pImgCols := int(math.Ceil(math.Sqrt(float64(*popSize))))
	pImgRows := int(math.Ceil(float64(*popSize) / float64(pImgCols)))
	imgTempPop := imgut.Create(pImgCols*settings.ImgTarget.W, pImgRows*settings.ImgTarget.H, settings.ImgTarget.ColorSpace)

	// Define the operators
	settings.CrossOver = makeCrossover(&settings)
	settings.Mutate = makeMultiMutation(&settings, *fMultiMut, *fMutSin, *fMutNod, *fMutSub, *fMutAre, *fMutLoc)

	// Seed rng
	if !*quiet {
		fmt.Println("Number of CPUs", runtime.NumCPU())
		runtime.GOMAXPROCS(runtime.NumCPU())
		fmt.Println("CPUs limits", runtime.GOMAXPROCS(0))
	}

	// Build population
	pop := new(base.Population)
	pop.Set = &settings
	pop.TournSize = *tournSize
	pop.Initialize(*popSize)

	// Save initial population FIXME it's for debugging
	/*
		for i := range pop.Pop {
			pop.Pop[i].Draw(imgTemp)
			imgTemp.WritePNG(fmt.Sprintf("pop_ind_%v.png", i))
		}
	*/

	// Number of parallel generators to setup
	pipelineSize := 1 // Was 4 FIXME but there are statistics used in parallel and not ready for concurrent access
	// Containers for pipelined operators
	chXo := make([]<-chan ga.PipelineIndividual, pipelineSize)
	chMut := make([]<-chan ga.PipelineIndividual, pipelineSize)

	// Save best individual, for elitism
	var elite ga.Individual = nil

	// Save time before starting
	//genTime := time.Now()

	// Loop until max number of generation is reached
	for g := 0; g < *numGen; g++ {
		// Compute fitness for every individual with no fitness
		fitnessEval := pop.Evaluate()
		fitnessEval = fitnessEval
		//
		//if !*quiet {
		//	fmt.Println("Generation ", g, "fit evals", fitnessEval, time.Since(genTime))
		//	genTime = time.Now()
		//}

		// Compute various statistics
		sta.Observe(pop)

		// Statistics and samples
		if g%*saveInterval == 0 {
			snapName, snapPopName := sta.SaveSnapshot(pop, *quiet, countersKeys, statsKeys, intCountersKeys)
			// Save best individual
			pop.BestIndividual().Fitness()
			pop.BestIndividual().(*base.Individual).ImgTemp.WritePNG(snapName)
			// Save best individual code

			//	pop.BestIndividual().(*base.Individual).Draw(settings.ImgTemp)
			//	settings.ImgTemp.WritePNG(snapName)
			// Save pop images
			pop.Draw(imgTempPop, pImgCols, pImgRows)
			imgTempPop.WritePNG(snapPopName)
		}

		// Setup parallel pipeline
		selectionSize := len(pop.Pop) // int(float64(len(pop.Pop))*0.3)) if you want to randomly generate new individuals
		chSel := ga.GenSelect(pop, selectionSize, float32(g)/float32(*numGen), elite)
		for i := 0; i < pipelineSize; i++ {
			chXo[i] = ga.GenCrossover(chSel, *pCross)
			chMut[i] = ga.GenMutate(chXo[i], *pMut)
		}
		var sel []ga.PipelineIndividual = ga.Collector(ga.FanIn(chMut...), selectionSize)

		// Replace old population and compute statistics
		for i := range sel {
			pop.Pop[i] = sel[i].Ind.(*base.Individual)
			sta.ObserveCrossoverFitness(sel[i].CrossoverFitness, sel[i].InitialFitness)
			sta.ObserveMutationFitness(sel[i].MutationFitness, sel[i].CrossoverFitness)
		}

		// When elitism is activated, get best individual
		if *fElite {
			elite = pop.BestIndividual()
		}

		// Build new individuals
		//base.RampedFill(pop, len(sel), len(pop.Pop))
	}
	fitnessEval := pop.Evaluate()
	fitnessEval = fitnessEval
	// Population statistics
	sta.Observe(pop)

	//if !*quiet {
	//	fmt.Println("Generation", *numGen, "fit evals", fitnessEval)
	//	fmt.Println("Best individual", pop.BestIndividual())
	//}

	snapName, snapPopName := sta.SaveSnapshot(pop, *quiet, countersKeys, statsKeys, intCountersKeys)
	// Save best individual
	pop.BestIndividual().Fitness()
	pop.BestIndividual().(*base.Individual).ImgTemp.WritePNG(snapName)
	//	pop.BestIndividual().(*base.Individual).Draw(settings.ImgTemp)
	//	settings.ImgTemp.WritePNG(snapName)
	// Save pop images
	pop.Draw(imgTempPop, pImgCols, pImgRows)
	imgTempPop.WritePNG(snapPopName)

	if !*quiet {
		fmt.Println("Best individual:")
		fmt.Println(pop.BestIndividual())
	}

	elapsedTime := time.Since(startTime)
	fmt.Println("Execution took %s", elapsedTime)

	/*
		bestName := fmt.Sprintf("%v/best/%v.png", basedir, basename)
		if !*quiet {
			fmt.Println("Saving best individual in", bestName)
			fmt.Println(pop.BestIndividual())
		}
		pop.BestIndividual().(*base.Individual).Draw(settings.ImgTemp)
		settings.ImgTemp.WritePNG(bestName)
	*/
}
