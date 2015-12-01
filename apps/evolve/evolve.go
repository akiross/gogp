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
	"math"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof" // profiling...
	"time"
)

func Evolve(calcMaxDepth func(*imgut.Image) int, fun, ter []gp.Primitive, drawfun func(*base.Individual, *imgut.Image)) {
	// Build settings
	var settings base.Settings
	// Primitives to use
	settings.Functionals = fun
	settings.Terminals = ter
	// Draw function to use
	settings.Draw = drawfun

	// Setup options
	seed := flag.Int64("seed", time.Now().UTC().UnixNano(), "Seed for RNG")
	numGen := flag.Int("g", 100, "Number of generations")
	popSize := flag.Int("p", 1000, "Size of population")
	saveInterval := flag.Int("n", 25, "Generations interval between two snapshot saves")
	tournSize := flag.Int("T", 3, "Tournament size")
	pCross := flag.Float64("C", 0.8, "Crossover probability")
	pMut := flag.Float64("M", 0.1, "Bit mutation probability")
	quiet := flag.Bool("q", false, "Quiet mode")
	//advStats := flag.Bool("stats", false, "Enable advanced statistics")
	//nps := flag.Bool("nps", false, "Disable population snapshot (no-pop-snap)")
	targetPath := flag.String("t", "", "Target image (PNG) path")
	var basedir, basename string
	cpuProfile := flag.String("cpuprofile", "", "Write CPU profile to file")

	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "  basedir (string) to be used for saving logs and files\n")
		fmt.Fprintf(os.Stderr, "  basename (string) to be used for saving logs and files\n")
	}

	flag.Parse()

	// Check if the argument
	args := flag.Args()

	if len(args) != 2 {
		flag.Usage()
		fmt.Fprintf(os.Stderr, "\nBasename/basedir parameter not specified\n")
		return
	} else {
		basedir = args[0]
		basename = args[1]
	}

	sta := stats.Create(basedir, basename)

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
		fmt.Println("ERROR: Cannot load image", *targetPath)
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
	settings.CrossOver = node.MakeTree1pCrossover(settings.MaxDepth)
	settings.Mutate = node.MakeTreeSingleMutation(settings.Functionals, settings.Terminals)

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

	// Loop until max number of generation is reached
	for g := 0; g < *numGen; g++ {
		// Compute fitness for every individual with no fitness
		fitnessEval := pop.Evaluate()
		if !*quiet {
			fmt.Println("Generation ", g, "fit evals", fitnessEval)
		}

		// Compute various statistics
		sta.Observe(pop)

		// Statistics and samples
		if g%*saveInterval == 0 {
			snapName, snapPopName := sta.SaveSnapshot(pop, *quiet)
			// Save best individual
			pop.BestIndividual().Evaluate()
			pop.BestIndividual().(*base.Individual).ImgTemp.WritePNG(snapName)
			//	pop.BestIndividual().(*base.Individual).Draw(settings.ImgTemp)
			//	settings.ImgTemp.WritePNG(snapName)
			// Save pop images
			pop.Draw(imgTempPop, pImgCols, pImgRows)
			imgTempPop.WritePNG(snapPopName)
		}

		// Setup parallel pipeline
		selectionSize := len(pop.Pop) // int(float64(len(pop.Pop))*0.3)) if you want to randomly generate new individuals
		chSel := ga.GenSelect(pop, selectionSize)
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
		// Build new individuals
		//base.RampedFill(pop, len(sel), len(pop.Pop))
	}
	fitnessEval := pop.Evaluate()
	// Population statistics
	sta.Observe(pop)

	if !*quiet {
		fmt.Println("Generation", *numGen, "fit evals", fitnessEval)
		fmt.Println("Best individual", pop.BestIndividual())
	}

	snapName, snapPopName := sta.SaveSnapshot(pop, *quiet)
	// Save best individual
	pop.BestIndividual().Evaluate()
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
