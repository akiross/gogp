package rr

import (
	"fmt"
	"github.com/akiross/gogp/gp"
	"github.com/akiross/gogp/image/draw2d/imgut"
	"github.com/akiross/gogp/node"
	"github.com/akiross/gogp/repr/rr"
	"math"
	"math/rand"
)

/***********************************
 * Genetic Operators
 **********************************/

// Define primitives
var Functionals []gp.Primitive = []gp.Primitive{rr.MakeHSplit(), rr.MakeVSplit()}
var Terminals []gp.Primitive = []gp.Primitive{}

func Draw(ind *node.Node, img *imgut.Image) {
	// We have to compile the nodes
	renderer := node.CompileTree(ind).(*rr.Primitive)
	// Apply the function
	renderer.Render(0, 0, float64(img.W), float64(img.H), img)
}

func MakeMaxDepth(limit int) func(*imgut.Image) int {
	return func(img *imgut.Image) int {
		// Compute the right value of maxDepth: each split divides image in 2
		// Hence 2^n = 1, 2, 4, 8, 16... is the number of splits we get at depth n
		// If the image has P pixels, we want to pick the smallest n such that 2^n > P -> n > log_2(P)
		md := int(math.Log2(float64(img.W*img.H))) + 1
		// Limit height to avoid too large trees
		if md > limit {
			return limit
		}
		return md
	}
}

func MakeFullColor() *rr.Primitive {
	c := rand.Float64()
	name := fmt.Sprintf("T_%d", int(c*255))
	return rr.MakeTerminal(name, rr.Filler(c, c, c))
}

func MakeShadeColor() *rr.Primitive {
	// Pick two random colors
	c, k := rand.Float64(), rand.Float64()
	// Pick random positions
	sx, sy, ex, ey := rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64()
	name := fmt.Sprintf("EPH_%x-%x_%d-%d_%d-%d", int(c*255), int(k*255), int(sx*100), int(sy*100), int(ex*100), int(ey*100))
	return rr.MakeTerminal(name, rr.LinShade(c, k, sx, sy, ex, ey))
}

func MakeDiagFill() *rr.Primitive {
	// Pick two random colors
	c, k := rand.Float64(), rand.Float64()
	// Pick a random diagonal
	d := rand.Intn(2) == 0
	var name string
	if d {
		name = fmt.Sprintf("Df_%x-%x", int(c*255), int(k*255))
	} else {
		name = fmt.Sprintf("dF_%x-%x", int(c*255), int(k*255))
	}
	return rr.MakeTerminal(name, rr.DiagShade(c, k, d))
}

func MakeDiagLine() *rr.Primitive {
	// Pick random back/foreground colors
	b, f := rand.Float64(), rand.Float64()
	// Pick a random diagonal
	d := rand.Intn(2) == 0
	// Pick a random line size
	s := rand.Intn(16)
	var name string
	if d {
		name = fmt.Sprintf("Dl_%x_%x-%x", int(s*15), int(b*255), int(f*255))
	} else {
		name = fmt.Sprintf("dL_%x_%x-%x", int(s*15), int(b*255), int(f*255))
	}
	return rr.MakeTerminal(name, rr.DiagLine(b, f, d, s))
}
