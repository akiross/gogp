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
	ImgTarget, ImgTemp *imgut.Image

	// Functionals and terminals used
	Functionals []gp.Primitive
	Terminals   []gp.Primitive

	Draw func(*Individual, *imgut.Image)

	// Operators used in evolution
	CrossOver func(float64, *node.Node, *node.Node)
	Mutate    func(float64, *node.Node)
}

type Individual struct {
	Node       *node.Node
	fitness    ga.Fitness
	fitIsValid bool
	set        *Settings
}

func (ind *Individual) String() string {
	return fmt.Sprint(ind.Node)
}

func (ind *Individual) Fitness() ga.Fitness {
	return ind.fitness
}

func (ind *Individual) Copy() ga.Individual {
	return &Individual{ind.Node.Copy(), ind.fitness, ind.fitIsValid, ind.set}
}

func (ind *Individual) Crossover(pCross float64, mate ga.Individual) {
	ind.set.CrossOver(pCross, ind.Node, mate.(*Individual).Node)
	ind.fitIsValid, mate.(*Individual).fitIsValid = false, false
}

func (ind *Individual) Draw(img *imgut.Image) {
	ind.set.Draw(ind, img)
}

func (ind *Individual) Evaluate() {
	// Draw the individual
	ind.set.Draw(ind, ind.set.ImgTemp)
	ind.fitness = ga.Fitness(imgut.PixelDistance(ind.set.ImgTemp, ind.set.ImgTarget))
	ind.fitIsValid = true
}

func (ind *Individual) FitnessValid() bool {
	return ind.fitIsValid
}

func (ind *Individual) Initialize() {
	ind.Node = node.MakeTreeHalfAndHalf(ind.set.MaxDepth, ind.set.Functionals, ind.set.Terminals)
}

// BUG(akiross) the mutation used here replaces a single, random.Node with an equivalent one - same as in DEAP - but we should go over each.Node and apply mutation probability
func (ind *Individual) Mutate(pMut float64) {
	ind.set.Mutate(pMut, ind.Node)
	ind.fitIsValid = false
}
