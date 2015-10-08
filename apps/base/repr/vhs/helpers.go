package vhs

import (
	"github.com/akiross/gogp/gp"
	"github.com/akiross/gogp/image/draw2d/imgut"
	"github.com/akiross/gogp/node"
	"github.com/akiross/gogp/repr/split/vhs"
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
var Functionals []gp.Primitive = []gp.Primitive{vhs.Functional(vhs.VSplit), vhs.Functional(vhs.HSplit)}
var Terminals []gp.Primitive = []gp.Primitive{vhs.Terminal(Black), vhs.Terminal(White)}

func Draw(ind *node.Node, img *imgut.Image) {
	// We have to compile the nodes
	exec := node.CompileTree(ind).(vhs.Terminal)
	// Apply the function
	exec(0, 0, float64(img.W), float64(img.H), img)
}

func MaxDepth(img *imgut.Image) int {
	// Compute the right value of maxDepth: each split divides image in 2
	// Hence 2^n = 1, 2, 4, 8, 16... is the number of splits we get at depth n
	// If the image has P pixels, we want to pick the smallest n such that 2^n > P -> n > log_2(P)
	md := int(math.Log2(float64(img.W*img.H))) + 1
	// Limit height to 7, to avoid too large trees
	if md > 7 {
		return 7
	}
	return md
}

func init() {
	// Build some colors
	count := 16
	for i := 1; i < count; i++ {
		c := float64(i) / float64(count)
		Terminals = append(Terminals, vhs.Terminal(vhs.Filler(c, c, c, 1)))
	}
}
