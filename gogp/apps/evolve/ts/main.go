package main

import (
	"github.com/akiross/gogp/apps/base"
	"github.com/akiross/gogp/apps/evolve"
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

func Gray(x1, x2, y float64, img *imgut.Image) {
	ts.Filler(0.5, 0.5, 0.5, 1)(x1, x2, y, img)
}

// Define primitives
var functionals []gp.Primitive = []gp.Primitive{ts.Functional(ts.Split)}

//var terminals []gp.Primitive = make([]gp.Primitive, 0, 10)
var terminals []gp.Primitive = []gp.Primitive{ts.Terminal(Black), ts.Terminal(White)}

func draw(ind *base.Individual, img *imgut.Image) {
	// We have to compile the nodes
	exec := node.CompileTree(ind.Node).(ts.Terminal)
	// Apply the function
	w := float64(img.W)
	exec(1.5*w, -0.5*w, 0, img)
}

func maxDepth(img *imgut.Image) int {
	// Compute the right value of maxDepth: each triangle splits in 4 parts the image
	// Hence 4^n = 1, 4, 16, 64, 256... is the number of splits we get at depth n
	// If the image has P pixels, we want to pick the smallest n such that 4^n > P -> n > log_2(P)/2
	md := int(math.Log2(float64(img.W*img.H))/2) + 1
	// Limit height to 4, to avoid trees with too many nodes
	if md > 4 {
		return 4
	}
	return md
}

func init() {
	// Build some colors
	/*
		count := 10
		for i := 0; i <= count; i++ {
			c := float64(i) / float64(count)
			terminals = append(terminals, ts.Terminal(ts.Filler(c, c, c, 1)))
		}
	*/
}

func main() {
	evolve.Evolve(maxDepth, functionals, terminals, draw)
}
