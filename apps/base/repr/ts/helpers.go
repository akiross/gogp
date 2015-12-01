package ts

import (
	"github.com/akiross/gogp/gp"
	"github.com/akiross/gogp/image/draw2d/imgut"
	"github.com/akiross/gogp/node"
	"github.com/akiross/gogp/repr/split/ts"
	"math"
)

/***********************************
 * Genetic Operators
 **********************************/

// type Terminal func(x1, x2, y float64, img *imgut.Image)

// Build terminals with names

func Black(x1, x2, y float64, img *imgut.Image) {
	ts.Filler(0, 0, 0, 1)(x1, x2, y, img)
}

func White(x1, x2, y float64, img *imgut.Image) {
	ts.Filler(1, 1, 1, 1)(x1, x2, y, img)
}

// Define primitives
var Functionals []gp.Primitive = []gp.Primitive{ts.Functional(ts.Split)}

//var terminals []gp.Primitive = make([]gp.Primitive, 0, 10)
var Terminals []gp.Primitive = []gp.Primitive{ts.Terminal(Black), ts.Terminal(White)}

func Draw(ind *node.Node, img *imgut.Image) {
	// We have to compile the nodes
	exec := node.CompileTree(ind).(ts.Terminal)
	// Apply the function
	w := float64(img.W)
	exec(1.5*w, -0.5*w, 0, img)
}

func MaxDepth(img *imgut.Image) int {
	// Compute the right value of maxDepth: each triangle splits in 4 parts the image
	// Hence 4^n = 1, 4, 16, 64, 256... is the number of splits we get at depth n
	// If the image has P pixels, we want to pick the smallest n such that 4^n > P -> n > log_2(P)/2
	md := int(math.Log2(float64(img.W*img.H))/2) + 1
	// Limit height to 4, to avoid trees with too many nodes
	if md > 10 {
		return 10
	}
	return md
}

func init() {
	// Build some colors
	count := 16
	for i := 1; i < count; i++ {
		c := float64(i) / float64(count)
		Terminals = append(Terminals, ts.Terminal(ts.Filler(c, c, c, 1)))
	}
}
