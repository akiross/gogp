package vhs

import (
	"fmt"
	"github.com/akiross/gogp/gp"
	"github.com/akiross/gogp/image/draw2d/imgut"
	"github.com/akiross/gogp/node"
	"github.com/akiross/gogp/repr/split/vhs"
	"math"
	"math/rand"
)

/***********************************
 * Genetic Operators
 **********************************/

// Define primitives
var Functionals []gp.Primitive = []gp.Primitive{vhs.Functional(vhs.VSplit), vhs.Functional(vhs.HSplit)}
var FuncNames []string = []string{"VSplit", "HSplit"}

var Terminals []gp.Primitive = []gp.Primitive{}
var TermNames []string = []string{}

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
	// Limit height to avoid too large trees
	if md > 13 {
		return 13
	}
	return md
}

// Buils a new ephimeral constant
func MakeEphimeral() gp.Primitive {
	sx, sy, ex, ey := rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64()
	Terminals = append(Terminals, vhs.Terminal(vhs.LinShade(c, k, sx, sy, ex, ey)))
	TermNames = append(TermNames, fmt.Sprintf("T_%d-%d_%d-%d_%d-%d", int(c*256), int(k*256), int(sx*100), int(sy*100), int(ex*100), int(ey*100)))
}

func init() {
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
}
