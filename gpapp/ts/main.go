package main

import (
	"ale-re.net/phd/gogp"
	"ale-re.net/phd/image/draw2d/imgut"
	"ale-re.net/phd/reprgp/split/ts"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof" // profiling...
	"sync"
	"time"
)

/***********************************
 * Genetic Operators
 **********************************/

var maxDepth int = 7 // Max allowed depth for trees

// type Terminal func(x1, x2, y float64, img *imgut.Image)

// Build terminals with names
func Black(x1, x2, y float64, img *imgut.Image) {
	ts.Filler(0, 0, 0, 1)(x1, x2, y, img)
}

func White(x1, x2, y float64, img *imgut.Image) {
	ts.Filler(1, 1, 1, 1)(x1, x2, y, img)
}

// Define primitives
var functionals []gogp.Primitive = []gogp.Primitive{ts.Functional(ts.Split)}
var terminals []gogp.Primitive = []gogp.Primitive{ts.Terminal(Black), ts.Terminal(White)}

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

func (ind *Individual) Crossover(pCross float64, mate *Individual) {
	if rand.Float64() >= pCross {
		return
	}
	crossOver(ind.node, mate.node)
	ind.fitIsValid, mate.fitIsValid = false, false
}

func (ind *Individual) Draw(img *imgut.Image) {
	// We have to compile the nodes
	exec := gogp.CompileTree(ind.node).(ts.Terminal)
	// Apply the function
	w := float64(img.W)
	exec(1.5*w, -0.5*w, 0, img)
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

func (pop *Population) BestIndividual() *Individual {
	return pop.best
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

func (pop *Population) Select(n int) ([]*Individual, error) {
	selectionSize, tournSize := n, pop.tournSize
	if (selectionSize < 1) || (tournSize < 1) {
		return nil, &ParamError{"Cannot have selectionSize < 1 or tournSize < 1"}
	}
	// Slice to store the new population
	newPop := make([]*Individual, selectionSize)
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

// This is a version of Select that is a stage in a pipeline. Will provide pointers to NEW individuals
func GenSelect(pop *Population, n int) <-chan *Individual {
	// Get sizes
	selectionSize, tournSize := n, pop.tournSize
	// A channel for output individuals
	out := make(chan *Individual, selectionSize)
	go func() {
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
			sel := &Individual{best.node.Copy(), best.fitness, best.fitIsValid}
			out <- sel
		}
		close(out)
	}()
	return out
}

func GenCrossover(in <-chan *Individual, pCross float64) <-chan *Individual {
	out := make(chan *Individual)
	go func() {
		// Continue forever
		for {
			// Take one item and if we got a real item
			i1, ok := <-in
			if ok {
				i2, ok := <-in
				if ok {
					// We got two items! Crossover
					i1.Crossover(pCross, i2)
					out <- i1
					out <- i2
				} else {
					// Can't crossover a single item
					out <- i1
				}
			} else {
				close(out)
				break
			}
		}
	}()
	return out
}

func GenMutate(in <-chan *Individual, pMut float64) <-chan *Individual {
	out := make(chan *Individual)
	go func() {
		for ind := range in {
			ind.Mutate(pMut)
			out <- ind
		}
		close(out)
	}()
	return out
}

func FanIn(in ...<-chan *Individual) <-chan *Individual {
	out := make(chan *Individual)
	// Used to wait that a set of goroutines finish
	var wg sync.WaitGroup
	// Function that copy from one channel to out
	emitter := func(c <-chan *Individual) {
		for i := range c {
			out <- i
		}
		wg.Done() // Signal the group
	}
	// How many goroutines to wait for
	wg.Add(len(in))
	// Start the routines
	for _, c := range in {
		go emitter(c)
	}

	// Wait for the emitters to finish, then close
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// Collects all the individuals from a channel and return a slice. Size is a hint for performances
func Collector(in <-chan *Individual, size int) []*Individual {
	pop := make([]*Individual, 0, size)
	for ind := range in {
		pop = append(pop, ind)
	}
	return pop
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
	_, _, _, _, _, _ = seed, numGen, popSize, tournSize, pCross, pMut

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

	// Compute the right value of maxDepth: each triangle splits in 4 parts the image
	// Hence 4^n = 1, 4, 16, 64, 256... is the number of splits we get at depth n
	// If the image has P pixels, we want to pick the smallest n such that 4^n > P -> n > log_2(P)/2
	logicalDepth := int(math.Log2(float64(imgTarget.W*imgTarget.H))/2) + 1
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
		var sel []*Individual
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
			chSel := GenSelect(pop, len(pop.pop))
			chXo1, chXo2, chXo3, chXo4 := GenCrossover(chSel, *pCross), GenCrossover(chSel, *pCross), GenCrossover(chSel, *pCross), GenCrossover(chSel, *pCross)
			chMut1, chMut2, chMut3, chMut4 := GenMutate(chXo1, *pMut), GenMutate(chXo2, *pMut), GenMutate(chXo3, *pMut), GenMutate(chXo4, *pMut)
			sel = Collector(FanIn(chMut1, chMut2, chMut3, chMut4), len(pop.pop))
		}

		// Update samples
		if g%*saveInterval == 0 {
			snapName := fmt.Sprintf("%v/snapshot/%v-snapshot-%v.png", basedir, basename, snapshot)
			if !*quiet {
				fmt.Println("Saving best individual snapshot", snapName)
				fmt.Println(pop.BestIndividual().node)
			}
			pop.BestIndividual().Draw(imgTemp)
			imgTemp.WritePNG(snapName)
			snapshot++
		}

		// Replace old population
		pop.pop = sel
	}
	fitnessEval := pop.Evaluate()

	if !*quiet {
		fmt.Println("Generation", *numGen, "fit evals", fitnessEval)
		fmt.Println("Best individual", pop.BestIndividual())
	}

	bestName := fmt.Sprintf("%v/best/%v.png", basedir, basename)
	if !*quiet {
		fmt.Println("Saving best individual in", bestName)
		fmt.Println(pop.BestIndividual().node)
	}
	pop.BestIndividual().Draw(imgTemp)
	imgTemp.WritePNG(bestName)
}
