package main

import (
	"ale-re.net/phd/gogp"
	"ale-re.net/phd/imgut"
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

const maxDepth int = 7 // Max allowed depth for trees

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
	// Get the byte data
	dataTemp := imgTemp.GetChannels()
	dataTarget := imgTarget.GetChannels()
	// If we are in G8 mode, we don't have to check all the channels
	var maxK int
	if imgTemp.ColorSpace == imgut.MODE_G8 {
		maxK = 1
	} else {
		maxK = len(dataTemp[0][0])
	}

	// Compute the distance
	var dist float64
	var count int
	for i := range dataTemp {
		for j := range dataTemp[i] {
			for k := 0; k < maxK; k++ {
				// In the python version this is not normalized, but I think it should be for precision issues
				diff := (float64(dataTemp[i][j][k]) - float64(dataTarget[i][j][k])) // / 255.0
				dist += diff * diff
				count++
			}
		}
	}
	// BUG(akiross) in the python version this is not normalized, but it should be!! For now, use un-normalized version to compare results, later fix this bug
	// ind.fitness = gogp.Fitness(math.Sqrt(dist / float64(count)))
	ind.fitness = gogp.Fitness(math.Sqrt(dist))
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
	imgTarget = imgut.Load(*targetPath)
	fmt.Println("Image format RGB?", imgTarget.ColorSpace == imgut.MODE_RGB, imgTarget.ColorSpace)
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

/*
func (ind *BoolIndividual10) Crossover(pCross float64, mate *BoolIndividual10) {
	// Crossover is probabilistic
	if rand.Float64() >= pCross {
		return
	}

	// Find crossover point (single)
	p := rand.Intn(len(ind.genotype)-1) + 1
	// Copy data after the point
	for i := p; i < len(ind.genotype); i++ {
		ind.genotype[i], mate.genotype[i] = mate.genotype[i], ind.genotype[i]
	}
	// Fitness is unknown and invalid, default
	ind.fitIsValid = false
	mate.fitIsValid = false
}

func (ind *BoolIndividual10) Evaluate() {
	ind.fitness = 0
	for _, g := range ind.genotype {
		if g {
			ind.fitness += 1
		}
	}
	ind.fitIsValid = true
}

func (ind *BoolIndividual10) FitnessValid() bool {
	return ind.fitIsValid
}

func (ind *BoolIndividual10) Initialize() {
	for i := 0; i < len(ind.genotype); i++ {
		ind.genotype[i] = rand.Intn(2) == 0
	}
}

func (ind *BoolIndividual10) Mutate(pMut float64) *BoolIndividual10 {
	var b BoolIndividual10
	// b fitness is invalid by default
	for i := 0; i < len(ind.genotype); i++ {
		if rand.Float64() < pMut {
			b.genotype[i] = !ind.genotype[i]
		} else {
			b.genotype[i] = ind.genotype[i]
		}
	}
	return &b
}

func (i *BoolIndividual10) String() string {
	s := "["
	for _, g := range i.genotype {
		if g {
			s += "1 "
		} else {
			s += "0 "
		}
	}
	return fmt.Sprint(s[0:len(s)-1]+"] = ", i.fitness)
}

func (pop *BoolPopulation10) Evaluate() (fitnessEval int) {
	for i := range pop.pop {
		var ind *BoolIndividual10 = &pop.pop[i]
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

func (pop *BoolPopulation10) Initialize(n int) {
	pop.pop = make([]BoolIndividual10, n)
	for i := range pop.pop {
		pop.pop[i].Initialize()
	}
}

// Tournament selection
func (pop *BoolPopulation10) Select(n int) ([]BoolIndividual10, error) {
	selectionSize, tournSize := n, pop.tournSize
	if (selectionSize < 1) || (tournSize < 1) {
		return nil, &ParamError{"Cannot have selectionSize < 1 or tournSize < 1"}
	}
	// Slice to store the new population
	newPop := make([]BoolIndividual10, selectionSize)
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
		newPop[i] = *best
	}
	return newPop, nil
}

func (pop *BoolPopulation10) BestIndividual() *BoolIndividual10 {
	return pop.best
}

func White(x1, x2, y1, y2 float64, img *imgut.Image) {
	gogp.RectFiller(1)(x1, x2, y1, y2, img)
}

func Gray(x1, x2, y1, y2 float64, img *imgut.Image) {
	gogp.RectFiller(0.5)(x1, x2, y1, y2, img)
}

func Black(x1, x2, y1, y2 float64, img *imgut.Image) {
	gogp.RectFiller(0)(x1, x2, y1, y2, img)
}

func main() {
	seed := flag.Int64("seed", time.Now().UTC().UnixNano(), "Seed for RNG")
	numGen := flag.Int("gen", 20, "Number of generations")
	popSize := flag.Int("pop", 50, "Size of population")
	tournSize := flag.Int("tourn", 3, "Tournament size")
	pCross := flag.Float64("pCross", 0.8, "Crossover probability")
	pMut := flag.Float64("pMut", 0.1, "Bit mutation probability")

	flag.Parse()

	fmt.Println("Seed used", *seed)
	rand.Seed(*seed)

	///

	// Create an image
	img := imgut.Create(200, 200, imgut.MODE_G8)
	// Fill the background
	//	img.FillRect(0, 0, float64(img.W), float64(img.H), 0.3)
	fSumXY := func(x, y int) float64 { return float64(x + y) }
	fSquaX := func(x, y int) float64 { return float64(x) * float64(x) }
	fSquaY := func(x, y int) float64 { return float64(y) * float64(y) }
	fConst := func(x, y int) float64 { return 3.14 }
	_, _, _, _ = fConst, fSquaX, fSquaY, fSumXY // To avoid complaining...
	img.FillMath(fSquaX)
	// Draw a couple of rects in the top-left part
	img.FillRect(10, 10, 50, 50, 1)
	img.FillRect(13, 14, 40, 40, 0)
	// Draw a couple of tris in the top-right part
	img.FillTriangle(130, 170, 50, 0)
	img.FillTriangle(190, 150, 40, 0.8)
	//img.DrawPoly(0, 0, 10, 10, 100, 10, 100, 100, 10, 100, 10, 10)
	fmt.Println("Stica", img)

	// Draw fractal rect split in bottom-left
	rTrue, rFalse := gogp.RectFiller(1), gogp.RectFiller(0)
	drawRect := gogp.VSplit(rTrue, gogp.HSplit(rFalse, gogp.VSplit(rFalse, rTrue)))
	drawRect(0, 100, 100, 200, img)

	// Draw fractal tri in bottom-right
	tTrue, tFalse := gogp.TriFiller(1), gogp.TriFiller(0)
	drawTri := gogp.TriSplit(tTrue, tTrue, gogp.TriSplit(tFalse, tFalse, tTrue, tFalse), tTrue)
	drawTri(100, 200, 200, img)

	// Get image data
	data := img.GetChannels(1)
	fmt.Printf("Type of channel data: %T %v %v %v\n", data, len(data), len(data[0]), len(data[0][0]))

	img.WritePNG("test_img.png")

	// Build some trees to test
	functionals, terminals := []gogp.Functional{gogp.VSplit, gogp.HSplit}, []gogp.Terminal{White, Gray, Black}

	gTree := gogp.MakeTreeGrow(4, 2, functionals, terminals)
	fmt.Println("Grown tree:", gTree)
	cTree := gogp.CompileTree(gTree)
	cTree(0, 0, float64(img.W), float64(img.H), img)

	img.WritePNG("test_img_grow.png")

	fTree := gogp.MakeTreeFull(4, 2, functionals, terminals)
	fmt.Println("Full tree:", fTree)
	cTree = gogp.CompileTree(fTree)
	cTree(0, 0, float64(img.W), float64(img.H), img)

	img.WritePNG("test_img_full.png")

	hTree := gogp.MakeTreeHalfAndHalf(4, 2, functionals, terminals)
	fmt.Println("HalfAndHalf tree:", hTree)
	cTree = gogp.CompileTree(hTree)
	cTree(0, 0, float64(img.W), float64(img.H), img)

	img.WritePNG("test_img_hnh.png")
	return

	// Create population
	pop := new(BoolPopulation10)
	pop.tournSize = *tournSize
	pop.Initialize(*popSize)

	// Save the best individual found so far
	//	var best *Individual = pop[0]

	// Loop until max number of generations is reached
	for g := 0; g < *numGen; g++ {
		fmt.Print("Generation ", g)

		// Compute fitness for every individual with no fitness
		fitnessEval := pop.Evaluate()
		fmt.Println(" fit evals", fitnessEval)

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

	// Evaluate the last changes
	fitnessEval := pop.Evaluate()
	fmt.Println("Generation", *numGen, "fit evals", fitnessEval)

	fmt.Println("Best individual", pop.BestIndividual())
}
*/
