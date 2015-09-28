package main

import (
	"ale-re.net/phd/gogp"
	"ale-re.net/phd/image/draw2d/imgut"
	"ale-re.net/phd/reprgp/split/vhs"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof" // profiling...
	"time"
)

/***********************************
 * Genetic Operators
 **********************************/

var maxDepth int = 7 // Max allowed depth for trees

// Build terminals with names
func Black(x1, y1, x2, y2 float64, img *imgut.Image) {
	vhs.Filler(0, 0, 0, 1)(x1, y1, x2, y2, img)
}

func White(x1, y1, x2, y2 float64, img *imgut.Image) {
	vhs.Filler(1, 1, 1, 1)(x1, y1, x2, y2, img)
}

// Define primitives
var functionals []gogp.Primitive = []gogp.Primitive{vhs.Functional(vhs.VSplit), vhs.Functional(vhs.HSplit)}
var terminals []gogp.Primitive = []gogp.Primitive{vhs.Terminal(Black), vhs.Terminal(White)}

// Images used for evaluation
var imgTarget, imgTemp *imgut.Image

// Operators used in evolution
var crossOver func(*gogp.Node, *gogp.Node)
var pointMutation func(*gogp.Node)

type Individual struct {
	node       *gogp.Node
	fitness    gogp.Fitness
	fitIsValid bool
}

type Population struct {
	best      *Individual
	pop       []*Individual
	tournSize int
	gogp.MinProblem
}

type ParamError struct {
	what string
}

func (e *ParamError) Error() string {
	return e.what
}

func (ind *Individual) Copy() gogp.Individual {
	return &Individual{ind.node.Copy(), ind.fitness, ind.fitIsValid}
}

func (ind *Individual) Crossover(pCross float64, mate gogp.Individual) {
	if rand.Float64() >= pCross {
		return
	}
	crossOver(ind.node, mate.(*Individual).node)
	ind.fitIsValid, mate.(*Individual).fitIsValid = false, false
}

func (ind *Individual) Draw(img *imgut.Image) {
	// We have to compile the nodes
	exec := gogp.CompileTree(ind.node).(vhs.Terminal)
	// Apply the function
	exec(0, 0, float64(img.W), float64(img.H), img)
}

func (ind *Individual) Evaluate() {
	// Draw the individual
	ind.Draw(imgTemp)
	ind.fitness = gogp.Fitness(imgut.PixelDistance(imgTemp, imgTarget))
	ind.fitIsValid = true
}

func (ind *Individual) FitnessValid() bool {
	return ind.fitIsValid
}

func (ind *Individual) Initialize() {
	ind.node = gogp.MakeTreeHalfAndHalf(maxDepth, functionals, terminals)
}

// BUG(akiross) the mutation used here replaces a single, random node with an equivalent one - same as in DEAP - but we should go over each node and apply mutation probability
func (ind *Individual) Mutate(pMut float64) {
	if rand.Float64() >= pMut {
		return
	}
	pointMutation(ind.node)
	ind.fitIsValid = false
}

func (ind *Individual) String() string {
	return fmt.Sprint(ind.node)
}

func (pop *Population) BestIndividual() gogp.Individual {
	return pop.best
}

func (pop *Population) Get(i int) gogp.Individual {
	return pop.pop[i]
}
func (pop *Population) Size() int {
	return len(pop.pop)
}

// Evaluate population fitness, return number of evaluations
func (pop *Population) Evaluate() (fitnessEval int) {
	for i := range pop.pop {
		var ind *Individual = pop.pop[i]
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
	pop.pop = make([]*Individual, n)
	for i := range pop.pop {
		pop.pop[i] = new(Individual)
		pop.pop[i].Initialize()
	}
}

func (pop *Population) Select(n int) ([]gogp.Individual, error) {
	selectionSize, tournSize := n, pop.tournSize
	if (selectionSize < 1) || (tournSize < 1) {
		return nil, &ParamError{"Cannot have selectionSize < 1 or tournSize < 1"}
	}
	// Slice to store the new population
	newPop := make([]gogp.Individual, selectionSize)
	for i := 0; i < selectionSize; i++ {
		// Pick an initial (pointer to) random individual
		best := pop.pop[rand.Intn(len(pop.pop))]
		// Select other players and select the best
		for j := 1; j < tournSize; j++ {
			maybe := pop.pop[rand.Intn(len(pop.pop))]
			if pop.BetterThan(maybe.fitness, best.fitness) {
				best = maybe
			}
		}
		newPop[i] = &Individual{best.node.Copy(), best.fitness, best.fitIsValid}
	}
	return newPop, nil
}

func main() {
	// Setup options
	seed := flag.Int64("seed", time.Now().UTC().UnixNano(), "Seed for RNG")
	numGen := flag.Int("g", 100, "Number of generations")
	popSize := flag.Int("p", 1000, "Size of population")
	saveInterval := flag.Int("n", 25, "Generations interval between two snapshot saves")
	tournSize := flag.Int("T", 3, "Tournament size")
	pCross := flag.Float64("C", 0.8, "Crossover probability")
	pMut := flag.Float64("M", 0.1, "Bit mutation probability")
	quiet := flag.Bool("q", false, "Quiet mode")
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
	imgTarget, err = imgut.Load(*targetPath)
	if err != nil {
		fmt.Println("ERROR: Cannot load image", *targetPath)
		panic("Cannot load image")
	}
	if !*quiet {
		fmt.Println("Image format RGB?", imgTarget.ColorSpace == imgut.MODE_RGB, imgTarget.ColorSpace)
	}

	// Compute the right value of maxDepth: each split divides image in 2
	// Hence 2^n = 1, 2, 4, 8, 16... is the number of splits we get at depth n
	// If the image has P pixels, we want to pick the smallest n such that 2^n > P -> n > log_2(P)
	logicalDepth := int(math.Log2(float64(imgTarget.W*imgTarget.H))) + 1
	if logicalDepth < maxDepth {
		maxDepth = logicalDepth
	}
	if !*quiet {
		fmt.Println("For area of", imgTarget.W*imgTarget.H, "pixels, max depth is", maxDepth)
	}

	// Create temporary surface, of same size and mode
	imgTemp = imgut.Create(imgTarget.W, imgTarget.H, imgTarget.ColorSpace)

	// Define the operators
	crossOver = gogp.MakeTree1pCrossover(maxDepth)
	pointMutation = gogp.MakeTreeNodeMutation(functionals, terminals)

	// Seed rng
	if !*quiet {
		fmt.Println("Seed used", *seed)
		fmt.Println("Number of CPUs", runtime.NumCPU())
		runtime.GOMAXPROCS(runtime.NumCPU())
		fmt.Println("CPUs limits", runtime.GOMAXPROCS(0))
	}
	rand.Seed(*seed)

	// Build population
	pop := new(Population)
	pop.tournSize = *tournSize
	pop.Initialize(*popSize)

	// Save initial population FIXME it's for debugging
	/*
		for i := range pop.pop {
			pop.pop[i].Draw(imgTemp)
			imgTemp.WritePNG(fmt.Sprintf("pop_ind_%v.png", i))
		}
	*/

	// Loop until max number of generation is reached
	snapshot := 0
	for g := 0; g < *numGen; g++ {

		// Compute fitness for every individual with no fitness
		fitnessEval := pop.Evaluate()
		if !*quiet {
			fmt.Println("Generation ", g, "fit evals", fitnessEval)
		}

		// set to true to use non-pipelined version
		var sel []gogp.Individual
		if false {
			// Apply selection
			sel, _ = pop.Select(len(pop.pop))

			// Crossover and mutation
			for i := 0; i < len(sel)-1; i += 2 {
				sel[i].Crossover(*pCross, sel[i+1])
				sel[i].Mutate(*pMut)
				sel[i+1].Mutate(*pMut)
			}

		} else {
			chSel := gogp.GenSelect(pop, len(pop.pop))
			chXo1, chXo2, chXo3, chXo4 := gogp.GenCrossover(chSel, *pCross), gogp.GenCrossover(chSel, *pCross), gogp.GenCrossover(chSel, *pCross), gogp.GenCrossover(chSel, *pCross)
			chMut1, chMut2, chMut3, chMut4 := gogp.GenMutate(chXo1, *pMut), gogp.GenMutate(chXo2, *pMut), gogp.GenMutate(chXo3, *pMut), gogp.GenMutate(chXo4, *pMut)
			sel = gogp.Collector(gogp.FanIn(chMut1, chMut2, chMut3, chMut4), len(pop.pop))
		}

		// Update samples
		if g%*saveInterval == 0 {
			snapName := fmt.Sprintf("%v/snapshot/%v-snapshot-%v.png", basedir, basename, snapshot)
			if !*quiet {
				fmt.Println("Saving best individual snapshot", snapName)
				fmt.Println(pop.BestIndividual())
			}
			pop.BestIndividual().(*Individual).Draw(imgTemp)
			imgTemp.WritePNG(snapName)
			snapshot++
		}

		// Replace old population
		for i := range sel {
			pop.pop[i] = sel[i].(*Individual)
		}
	}
	fitnessEval := pop.Evaluate()

	if !*quiet {
		fmt.Println("Generation", *numGen, "fit evals", fitnessEval)
		fmt.Println("Best individual", pop.BestIndividual())
	}

	bestName := fmt.Sprintf("%v/best/%v.png", basedir, basename)
	if !*quiet {
		fmt.Println("Saving best individual in", bestName)
		fmt.Println(pop.BestIndividual())
	}
	pop.BestIndividual().(*Individual).Draw(imgTemp)
	imgTemp.WritePNG(bestName)
}
