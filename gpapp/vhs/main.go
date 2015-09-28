package main

import (
	"ale-re.net/phd/gogp"
	"ale-re.net/phd/gpapp"
	"ale-re.net/phd/image/draw2d/imgut"
	"ale-re.net/phd/reprgp/split/vhs"
	"math"
)

/***********************************
 * Genetic Operators
 **********************************/

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

func draw(ind *gpapp.Individual, img *imgut.Image) {
	// We have to compile the nodes
	exec := gogp.CompileTree(ind.Node).(vhs.Terminal)
	// Apply the function
	exec(0, 0, float64(img.W), float64(img.H), img)
}

func maxDepth(img *imgut.Image) int {
	// Compute the right value of maxDepth: each split divides image in 2
	// Hence 2^n = 1, 2, 4, 8, 16... is the number of splits we get at depth n
	// If the image has P pixels, we want to pick the smallest n such that 2^n > P -> n > log_2(P)
	return int(math.Log2(float64(img.W*img.H))) + 1
}

func main() {
	gpapp.Evolve(maxDepth, functionals, terminals, draw)
}
