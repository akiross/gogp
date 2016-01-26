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
var Terminals []gp.Primitive = []gp.Primitive{rr.MakeEphimeral("MakeFull", MakeFullColor), rr.MakeEphimeral("MakeShade", MakeShadeColor)}

func Draw(ind *node.Node, img *imgut.Image) {
	// We have to compile the nodes
	renderer := node.CompileTree(ind).(*rr.Primitive)
	// Apply the function
	renderer.Render(0, 0, float64(img.W), float64(img.H), img)
}

func MaxDepth(img *imgut.Image) int {
	// Compute the right value of maxDepth: each split divides image in 2
	// Hence 2^n = 1, 2, 4, 8, 16... is the number of splits we get at depth n
	// If the image has P pixels, we want to pick the smallest n such that 2^n > P -> n > log_2(P)
	md := int(math.Log2(float64(img.W*img.H))) + 1
	// Limit height to avoid too large trees
	if md > 13 {
		return 13
	}
	return md
}

func MakeFullColor() *rr.Primitive {
	c := rand.Float64()
	name := fmt.Sprintf("T_%d", int(c*256))
	return rr.MakeTerminal(name, rr.Filler(c, c, c))
}

func MakeShadeColor() *rr.Primitive {
	// Pick two random colors
	c, k := rand.Float64(), rand.Float64()
	// Pick random positions
	sx, sy, ex, ey := rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64()
	name := fmt.Sprintf("EPH_%x-%x_%d-%d_%d-%d", int(c*256), int(k*256), int(sx*100), int(sy*100), int(ex*100), int(ey*100))
	return rr.MakeTerminal(name, rr.LinShade(c, k, sx, sy, ex, ey))
}

func init() {
	//	Terminals = append(Terminals, rr.MakeTerminal("White", rr.Filler(1.0, 1.0, 1.0)))
	/*
		// Build some colors
		count := 8 // number of total colors, from black to white
		for i := 0; i <= count; i++ {
			c := float64(i) / float64(count)
			Terminals = append(Terminals, vhs.Terminal(vhs.Filler(c, c, c, 1)))
			TermNames = append(TermNames, fmt.Sprintf("T_%d", int(c*256)))
		}

		count = 8
		reps := 8
		for i := 0; i <= count; i++ {
			c := float64(i) / float64(count)
			for j := i + 1; j <= count; j++ {
				k := float64(j) / float64(count)
				// Multiple copie
				for n := 0; n < reps; n++ {
					sx, sy, ex, ey := rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64()
					Terminals = append(Terminals, vhs.Terminal(vhs.LinShade(c, k, sx, sy, ex, ey)))
					TermNames = append(TermNames, fmt.Sprintf("T_%d-%d_%d-%d_%d-%d", int(c*256), int(k*256), int(sx*100), int(sy*100), int(ex*100), int(ey*100)))
				}
			}
		}
	*/
}
