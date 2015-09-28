package main

import (
	"ale-re.net/phd/gogp"
	"ale-re.net/phd/gpapp"
	"ale-re.net/phd/image/draw2d/imgut"
	"ale-re.net/phd/reprgp/expr/binary"
	"math"
)

/***********************************
 * Genetic Operators
 **********************************/

// Define primitives
var functionals []gogp.Primitive = []gogp.Primitive{
	binary.Functional2(binary.Sum),
	binary.Functional2(binary.Sub),
	binary.Functional2(binary.Mul),
	binary.Functional2(binary.ProtectedDiv),
	binary.Functional1(binary.Square),
	binary.Functional1(binary.Abs),
}

var terminals []gogp.Primitive = []gogp.Primitive{
	binary.Terminal(binary.IdentityX),
	binary.Terminal(binary.IdentityY),
	binary.Terminal(binary.Constant(-10)),
	binary.Terminal(binary.Constant(-5)),
	binary.Terminal(binary.Constant(-2)),
	binary.Terminal(binary.Constant(-1)),
	binary.Terminal(binary.Constant(0)),
	binary.Terminal(binary.Constant(1)),
	binary.Terminal(binary.Constant(2)),
	binary.Terminal(binary.Constant(5)),
	binary.Terminal(binary.Constant(10)),
}

func draw(ind *gpapp.Individual, img *imgut.Image) {
	// We have to compile the nodes
	exec := gogp.CompileTree(ind.Node).(binary.Terminal)
	// Apply the function
	//	exec(0 0, float64(img.W), float64(img.H), img)
	var call imgut.PixelFunc = func(x, y int) float64 {
		return float64(exec(binary.NumericIn(x), binary.NumericIn(y)))
	}
	img.FillMath(call)
}

func maxDepth(img *imgut.Image) int {
	// Compute the right value of maxDepth: each triangle splits in 4 parts the image
	// Hence 4^n = 1, 4, 16, 64, 256... is the number of splits we get at depth n
	// If the image has P pixels, we want to pick the smallest n such that 4^n > P -> n > log_2(P)/2
	return int(math.Log2(float64(img.W*img.H))/2) + 1
}

func main() {
	gpapp.Evolve(maxDepth, functionals, terminals, draw)
}
