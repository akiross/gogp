package main

import (
	"ale-re.net/phd/gogp"
	"ale-re.net/phd/image/draw2d/imgut"
	"ale-re.net/phd/reprgp/split/vhs"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime/pprof" // profiling...
	"time"
)

/***********************************
 * Genetic Operators
 **********************************/
/*
type Individual interface {
	Crossover(float64, *Individual) // Crossover with given probability
	Evaluate()                      // Evaluate and cache fitness
	FitnessValid() bool             // True if fitness is valid
	Initialize()                    // Be initializable
	Mutate(p float64)               // Mutate with given probability
	fmt.Stringer                    // Be convertible to string
}

type Population interface {
	Evaluate() int                      // Evaluate fitnesses, return evaluated ones
	Initialize(n int)                   // Build N individuals
	Select(n int) ([]Individual, error) // Select N individuals
	BestIndividual() *Individual        // Return the best individual in current population
}
*/

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
	pop       []Individual
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

func (pop *Population) BestIndividual() *Individual {
	return pop.best
}

// Evaluate population fitness, return number of evaluations
func (pop *Population) Evaluate() (fitnessEval int) {
	for i := range pop.pop {
		var ind *Individual = &pop.pop[i]
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
	pop.pop = make([]Individual, n)
	for i := range pop.pop {
		pop.pop[i].Initialize()
	}
}

func (pop *Population) Select(n int) ([]Individual, error) {
	selectionSize, tournSize := n, pop.tournSize
	if (selectionSize < 1) || (tournSize < 1) {
		return nil, &ParamError{"Cannot have selectionSize < 1 or tournSize < 1"}
	}
	// Slice to store the new population
	newPop := make([]Individual, selectionSize)
	// Perform tournaments
	for i := 0; i < selectionSize; i++ {
		// Pick an initial (pointer to) random individual
		best := &pop.pop[rand.Intn(len(pop.pop))]
		// Select other players and select the best
		for j := 1; j < tournSize; j++ {
			maybe := &pop.pop[rand.Intn(len(pop.pop))]
			if pop.BetterThan(maybe.fitness, best.fitness) {
				best = maybe
			}
		}
		// Save winner (copy it)
		newPop[i] = Individual{best.node.Copy(), best.fitness, best.fitIsValid}
	}
	return newPop, nil
}

func main() {
	// Setup options
	seed := flag.Int64("seed", time.Now().UTC().UnixNano(), "Seed for RNG")
	numGen := flag.Int("gen", 100, "Number of generations")
	popSize := flag.Int("pop", 1000, "Size of population")
	tournSize := flag.Int("tourn", 3, "Tournament size")
	pCross := flag.Float64("pCross", 0.8, "Crossover probability")
	pMut := flag.Float64("pMut", 0.1, "Bit mutation probability")
	targetPath := flag.String("target", "", "Target image (PNG) path")

	cpuProfile := flag.String("cpuprofile", "", "Write CPU profile to file")

	flag.Parse()
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
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
	fmt.Println("Image format RGB?", imgTarget.ColorSpace == imgut.MODE_RGB, imgTarget.ColorSpace)

	// Compute the right value of maxDepth: each split divides image in 2
	// Hence 2^n = 1, 2, 4, 8, 16... is the number of splits we get at depth n
	// If the image has P pixels, we want to pick the smallest n such that 2^n > P -> n > log_2(P)
	logicalDepth := int(math.Log2(float64(imgTarget.W*imgTarget.H))) + 1
	if logicalDepth < maxDepth {
		maxDepth = logicalDepth
	}
	fmt.Println("For area of", imgTarget.W*imgTarget.H, "pixels, max depth is", maxDepth)

	// Create temporary surface, of same size and mode
	imgTemp = imgut.Create(imgTarget.W, imgTarget.H, imgTarget.ColorSpace)

	// Define the operators
	crossOver = gogp.MakeTree1pCrossover(maxDepth)
	pointMutation = gogp.MakeTreeNodeMutation(functionals, terminals)

	// Seed rng
	fmt.Println("Seed used", *seed)
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
	for g := 0; g < *numGen; g++ {
		fmt.Print("Generation ", g)

		// Compute fitness for every individual with no fitness
		fitnessEval := pop.Evaluate()
		fmt.Println(" fit evals", fitnessEval)

		// BUG(akiross) the following could be faster by doing 1 loop for sel, xo, mut

		// Apply selection
		sel, _ := pop.Select(len(pop.pop))

		// Crossover and mutation
		for i := 0; i < len(sel)-1; i += 2 {
			sel[i].Crossover(*pCross, &sel[i+1])
			sel[i].Mutate(*pMut)
			sel[i+1].Mutate(*pMut)
		}

		// Replace old population
		pop.pop = sel
	}
	fitnessEval := pop.Evaluate()
	fmt.Println("Generation", *numGen, "fit evals", fitnessEval)

	fmt.Println("Best individual", pop.BestIndividual())
	pop.BestIndividual().Draw(imgTemp)
	imgTemp.WritePNG("best-individual.png")
}
