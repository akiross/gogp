package base

import (
	"fmt"
	"github.com/akiross/gogp/ga"
	"github.com/akiross/gogp/gp"
	"github.com/akiross/gogp/image/draw2d/imgut"
	"github.com/akiross/gogp/node"
)

type Settings struct {
	// We keep this low, because trees may grow too large
	// and use too much memory
	MaxDepth int

	// Images used for evaluation
	//ImgTarget, ImgTemp *imgut.Image
	ImgTarget *imgut.Image

	// Functionals and terminals used
	Functionals []gp.Primitive
	Terminals   []gp.Primitive

	Draw func(*Individual, *imgut.Image)

	// Operators used in evolution
	CrossOver func(float64, *node.Node, *node.Node) bool
	Mutate    func(float64, *node.Node) bool
}

type Individual struct {
	Node       *node.Node
	fitness    ga.Fitness
	fitIsValid bool
	set        *Settings
	ImgTemp    *imgut.Image // where to render the individual
}

func (ind *Individual) String() string {
	return fmt.Sprint(ind.Node)
}

func (ind *Individual) Fitness() ga.Fitness {
	return ind.fitness
}

func (ind *Individual) Copy() ga.Individual {
	tmpImg := imgut.Create(ind.set.ImgTarget.W, ind.set.ImgTarget.H, ind.set.ImgTarget.ColorSpace)
	return &Individual{ind.Node.Copy(), ind.fitness, ind.fitIsValid, ind.set, tmpImg}
}

func (ind *Individual) Crossover(pCross float64, mate ga.Individual) {
	if ind.set.CrossOver(pCross, ind.Node, mate.(*Individual).Node) {
		ind.fitIsValid, mate.(*Individual).fitIsValid = false, false
	}
}

func (ind *Individual) Draw(img *imgut.Image) {
	ind.set.Draw(ind, img)
}

func (ind *Individual) Evaluate() ga.Fitness {
	// Compute only if necessary
	if true || !ind.fitIsValid { // BUG(akiross) FIXME TODO this test should be enabled
		// Draw the individual
		ind.set.Draw(ind, ind.ImgTemp)
		ind.fitness = ga.Fitness(imgut.PixelDistance(ind.ImgTemp, ind.set.ImgTarget))
		ind.fitIsValid = true
	}
	return ind.fitness
}

func (ind *Individual) FitnessValid() bool {
	return ind.fitIsValid
}

func (ind *Individual) Initialize() {
	ind.Node = node.MakeTreeHalfAndHalf(ind.set.MaxDepth, ind.set.Functionals, ind.set.Terminals)
	ind.ImgTemp = imgut.Create(ind.set.ImgTarget.W, ind.set.ImgTarget.H, ind.set.ImgTarget.ColorSpace)
}

// BUG(akiross) the mutation used here replaces a single, random.Node with an equivalent one - same as in DEAP - but we should go over each.Node and apply mutation probability
func (ind *Individual) Mutate(pMut float64) {
	if ind.set.Mutate(pMut, ind.Node) {
		ind.fitIsValid = false
	}
}
