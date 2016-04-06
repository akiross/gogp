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

func makeMultiMutation(s *base.Settings, enbSin, enbNod, enbSub, enbAre, enbLoc bool) func(float64, *base.Individual) bool {
	genFun := func(maxDep int) *node.Node {
		return node.MakeTreeHalfAndHalf(maxDep, s.Functionals, s.Terminals)
	}

	singleMut := node.MakeTreeSingleMutation(s.Functionals, s.Terminals) // func(*Node)
	nodeMut := node.MakeTreeNodeMutation(s.Functionals, s.Terminals)     // funcs, terms []gp.Primitive)// func(*Node) int {
	subtrMut := node.MakeSubtreeMutation(s.MaxDepth, genFun)
	areaMut := node.MakeSubtreeMutationGuided(s.MaxDepth, genFun, node.ArityDepthProbComputer)

	return func(pMut float64, ind *base.Individual) bool {
		// Generate a permutation of 5 numbers
		perm := rand.Perm(5)
		for _, v := range perm {
			// Pick the appropriate algorithm
			switch v {
			case 0:
				if enbSin {
					// Statistics on output values
					event := rand.Float64() < pMut
					ind.CountEvent("mut-single-event", event)
					if event {
						fit := ind.Evaluate()
						singleMut(ind.Node)
						newFit := ind.Evaluate()
						ind.CountEvent("mut-single-improv", s.BetterThan(newFit, fit))
					}
					return event
				}
			case 1:
				if enbNod {
					//fmt.Println("Node mutation")
					// pMut is applied to each node
					return nodeMut(pMut, ind.Node) != 0
				}
			case 2:
				if enbSub {
					//fmt.Println("Subtree mutation")
					event := rand.Float64() < pMut
					ind.CountEvent("mut-subtree-event", event)
					if event {
						fit := ind.Evaluate()
						subtrMut(ind.Node)
						newFit := ind.Evaluate()
						ind.CountEvent("mut-subtree-improv", s.BetterThan(newFit, fit))
					}
					return event
				}
			case 3:
				if enbAre {
					//fmt.Println("Area mutation")
					event := rand.Float64() < pMut
					ind.CountEvent("mut-area-event", event)
					if event {
						fit := ind.Evaluate()
						areaMut(ind.Node)
						newFit := ind.Evaluate()
						ind.CountEvent("mut-area-improv", s.BetterThan(newFit, fit))
					}
					return event
				}
			default:
				if enbLoc {
					//fmt.Println("Local search mutation")
					// Local search
					neighbSize := 5
					event := rand.Float64() < pMut
					ind.CountEvent("mut-local-event", event)
					if event {
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
						ind.CountEvent("mut-local-improv", s.BetterThan(newFit, fit))
					}
					return event
				}
			}
			// In the case we don't execute anything, go to the next method
		}
		// If this happens, all the mutations were disabled
		return false
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
	// Build settings
	var settings base.Settings
	// Primitives to use
	settings.Functionals = fun
	settings.Terminals = ter
	// Draw function to use
	settings.Draw = drawfun

	startTime := time.Now()

	// Setup options
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	seed := fs.Int64("seed", startTime.UTC().UnixNano(), "Seed for RNG")
	numGen := fs.Int("g", 100, "Number of generations")
	popSize := fs.Int("p", 1000, "Size of population")
	saveInterval := fs.Int("n", 25, "Generations interval between two snapshot saves")
	tournSize := fs.Int("T", 3, "Tournament size")
	pCross := fs.Float64("C", 0.8, "Crossover probability")
	pMut := fs.Float64("M", 0.1, "Bit mutation probability")
	quiet := fs.Bool("q", false, "Quiet mode")
	fElite := fs.Bool("el", false, "Enable elite individual")

	fMutSin := fs.Bool("ms", false, "Enable Single Mutation")
	fMutNod := fs.Bool("mn", false, "Enable Node Mutation")
	fMutSub := fs.Bool("mt", false, "Enable Subtree Mutation")
	fMutAre := fs.Bool("ma", false, "Enable Area Mutation")
	fMutLoc := fs.Bool("ml", false, "Enable Local Mutation")

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

	// Build statistic map
	settings.Statistics = make(map[string]*sequence.SequenceStats)
	settings.Counters = make(map[string]*counter.Counter)

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
	settings.Mutate = makeMultiMutation(&settings, *fMutSin, *fMutNod, *fMutSub, *fMutAre, *fMutLoc)

	// Seed rng
	if !*quiet {
		fmt.Println("Seed used", *seed)
		fmt.Println("Number of CPUs", runtime.NumCPU())
		runtime.GOMAXPROCS(runtime.NumCPU())
		fmt.Println("CPUs limits", runtime.GOMAXPROCS(0))
	}
	rand.Seed(*seed)

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
	pipelineSize := 4
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
			snapName, snapPopName := sta.SaveSnapshot(pop, *quiet)
			// Save best individual
			pop.BestIndividual().Fitness()
			pop.BestIndividual().(*base.Individual).ImgTemp.WritePNG(snapName)
			// Save best individual code

			//	pop.BestIndividual().(*base.Individual).Draw(settings.ImgTemp)
			//	settings.ImgTemp.WritePNG(snapName)
			// Save pop images
			pop.Draw(imgTempPop, pImgCols, pImgRows)
			imgTempPop.WritePNG(snapPopName)

			// Print custom statistics
			for key, sst := range settings.Statistics {
				fmt.Println("Stats", key, sst.Variance.Count(), sst.Min.Get(), sst.Max.Get(), sst.PartialMean(), sst.PartialVarBessel())
				sst.Clear()
			}
			for key, cst := range settings.Counters {
				fmt.Println("Count", key, cst.AbsoluteFrequency(), cst.RelativeFrequency())
				cst.Clear()
			}
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

	snapName, snapPopName := sta.SaveSnapshot(pop, *quiet)
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
