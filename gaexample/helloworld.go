package main

import (
	"ale-re.net/phd/gogp"
	"ale-re.net/phd/imgut"
	"flag"
	"fmt"
	"math/rand"
	"time"
)

/******************
 * Rect functions
 *****************/

// The function that will be called to get a solution
type RectTerminal func(x1, x2, y1, y2 float64, img *imgut.Image)

// Functionals are internal nodes of a solution tree
type RectFunctional func(args ...RectTerminal) RectTerminal

func (self RectTerminal) IsFunctional() bool {
	return false
}

func (self RectTerminal) Arity() int {
	return -1 // Not used
}

func (self RectTerminal) Run(p ...gogp.Primitive) gogp.Primitive {
	return self
}

func (self RectFunctional) IsFunctional() bool {
	return true
}

func (self RectFunctional) Arity() int {
	return 2
}

func (self RectFunctional) Run(p ...gogp.Primitive) gogp.Primitive {
	return self(p[0].(RectTerminal), p[1].(RectTerminal))
}

// Buils a Terminal that fills the entire area with given color
func RectFiller(col ...float64) RectTerminal {
	return func(x1, y1, x2, y2 float64, img *imgut.Image) {
		img.FillRect(x1, y1, x2, y2, col...)
	}
}

// Returns a terminal that fills the rectangle according to left and right
func VSplit(args ...RectTerminal) RectTerminal {
	return func(x1, y1, x2, y2 float64, img *imgut.Image) {
		xh := (x1 + x2) * 0.5
		args[0](x1, y1, xh, y2, img)
		args[1](xh, y1, x2, y2, img)
	}
}

func HSplit(args ...RectTerminal) RectTerminal {
	return func(x1, y1, x2, y2 float64, img *imgut.Image) {
		yh := (y1 + y2) * 0.5
		args[0](x1, y1, x2, yh, img)
		args[1](x1, yh, x2, y2, img)
	}
}

/*********************
 * Triangle functions
 ********************/

type TriTerminal func(x1, x2, y float64, img *imgut.Image)

type TriFunctional func(top, left, center, right TriTerminal) TriTerminal

// Return a terminal that fills the entire triangle with given color
func TriFiller(col ...float64) TriTerminal {
	return func(x1, x2, y float64, img *imgut.Image) {
		img.FillTriangle(x1, x2, y, col...)
	}
}

func TriSplit(top, left, center, right TriTerminal) TriTerminal {
	return func(x1, x2, y float64, img *imgut.Image) {
		// Split the triangle in 4 parts
		cx1, cxm, cx2 := x1+0.25*(x2-x1), x1+0.5*(x2-x1), x1+0.75*(x2-x1)
		ty := imgut.TriangleCenterY(x1, x2, y)
		cy := 0.5 * (ty + y)

		top(cx1, cx2, cy, img)
		left(x1, cxm, y, img)
		center(cx2, cx1, cy, img)
		right(cxm, x2, y, img)
	}
}

/***********************
 * Genetic algorithms
 **********************/

type ParamError struct {
	what string
}

func (e *ParamError) Error() string {
	return e.what
}

// A specific type of individual
type BoolIndividual10 struct {
	genotype   [10]bool
	fitness    gogp.Fitness
	fitIsValid bool
}

type BoolPopulation10 struct {
	best            *BoolIndividual10
	pop             []BoolIndividual10
	tournSize       int
	gogp.MinProblem // Minimization problem (promoted)
}

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
	RectFiller(1)(x1, x2, y1, y2, img)
}

func Gray(x1, x2, y1, y2 float64, img *imgut.Image) {
	RectFiller(0.5)(x1, x2, y1, y2, img)
}

func Black(x1, x2, y1, y2 float64, img *imgut.Image) {
	RectFiller(0)(x1, x2, y1, y2, img)
}

func main() {
	seed := flag.Int64("seed", time.Now().UTC().UnixNano(), "Seed for RNG")
	numGen := flag.Int("gen", 20, "Number of generations")
	popSize := flag.Int("pop", 50, "Size of population")
	tournSize := flag.Int("tourn", 3, "Tournament size")
	pCross := flag.Float64("pCross", 0.8, "Crossover probability")
	pMut := flag.Float64("pMut", 0.1, "Bit mutation probability")

	flag.Parse()

	//	*seed = 1442336044574955058
	//	*seed = 1442353043560344044 // a seed with a grow of height 2
	// *seed = 1442591264978496417 // Produces a wrong crossover

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
	rTrue, rFalse := RectFiller(1), RectFiller(0)
	drawRect := VSplit(rTrue, HSplit(rFalse, VSplit(rFalse, rTrue)))
	drawRect(0, 100, 100, 200, img)

	// Draw fractal tri in bottom-right
	tTrue, tFalse := TriFiller(1), TriFiller(0)
	drawTri := TriSplit(tTrue, tTrue, TriSplit(tFalse, tFalse, tTrue, tFalse), tTrue)
	drawTri(100, 200, 200, img)

	// Get image data
	data := img.GetChannels(1)
	fmt.Printf("Type of channel data: %T %v %v %v\n", data, len(data), len(data[0]), len(data[0][0]))

	img.WritePNG("test_img.png")

	// Build some trees to test
	functionals, terminals := []gogp.Primitive{RectFunctional(VSplit), RectFunctional(HSplit)}, []gogp.Primitive{RectTerminal(White), RectTerminal(Gray), RectTerminal(Black)}

	gTree := gogp.MakeTreeGrow(4, functionals, terminals)
	fmt.Println("Grown tree:\n", gTree.PrettyPrint())
	cTree := gogp.CompileTree(gTree).(RectTerminal)
	cTree(0, 0, float64(img.W), float64(img.H), img)

	enNodes, enDepth, enHeight := gTree.Enumerate()
	fmt.Println("-- Grown tree nodes, depth and heights", enNodes, enDepth, enHeight)

	img.WritePNG("test_img_grow.png")

	fTree := gogp.MakeTreeFull(3, functionals, terminals)
	fmt.Println("Full tree:\n", fTree.PrettyPrint())
	cTree = gogp.CompileTree(fTree).(RectTerminal)
	cTree(0, 0, float64(img.W), float64(img.H), img)

	enNodes, enDepth, enHeight = fTree.Enumerate()
	fmt.Println("-- Full tree nodes, depth and heights", enNodes, enDepth, enHeight)
	img.WritePNG("test_img_full.png")

	hTree := gogp.MakeTreeHalfAndHalf(4, functionals, terminals)
	fmt.Println("HalfAndHalf tree:\n", hTree.PrettyPrint())

	fmt.Println(hTree.GraphvizDot("hTree", hTree.Colorize(1, 0)))

	cTree = gogp.CompileTree(hTree).(RectTerminal)
	cTree(0, 0, float64(img.W), float64(img.H), img)

	img.WritePNG("test_img_hnh.png")

	fmt.Println("Testing mutation on hnh tree")
	mutation := gogp.MakeTreeNodeMutation(functionals, terminals)
	mutation(hTree)
	fmt.Println("Mutation 1\n", hTree.PrettyPrint())
	mutation(hTree)
	fmt.Println("Mutation 2\n", hTree.PrettyPrint())

	xoD4 := gogp.MakeTree1pCrossover(4)
	xoD5 := gogp.MakeTree1pCrossover(5)
	xoD6 := gogp.MakeTree1pCrossover(6)

	t1 := gogp.MakeTreeFull(4, functionals, terminals)
	t2 := gogp.MakeTreeFull(4, functionals, terminals)

	//enNodes, enDepth, enHeight = t1.Enumerate()
	//fmt.Println("-- Full tree 1 enums", enNodes, enDepth, enHeight)

	//enNodes, enDepth, enHeight = t1.Enumerate()
	//fmt.Println("-- Full tree 2 enums", enNodes, enDepth, enHeight)

	// Build a joint map of colors to be used for both threes
	var t1t2Cols map[*gogp.Node][]float32
	t1t2Cols = t1.Colorize(1.0, 0)
	for k, v := range t2.Colorize(0.5, 0) {
		t1t2Cols[k] = v
	}

	fmt.Println("Testing crossover with trees")
	t1.GraphvizDot("xoSt1", t1t2Cols)
	t2.GraphvizDot("xoSt2", t1t2Cols)
	xoD4(t1, t2)
	t1.GraphvizDot("xoD4t1", t1t2Cols)
	t2.GraphvizDot("xoD4t2", t1t2Cols)
	xoD5(t1, t2)
	t1.GraphvizDot("xoD5t1", t1t2Cols)
	t2.GraphvizDot("xoD5t2", t1t2Cols)
	xoD6(t1, t2)
	t1.GraphvizDot("xoD6t1", t1t2Cols)
	t2.GraphvizDot("xoD6t2", t1t2Cols)

	// Test crossover
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
