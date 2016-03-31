package expr

import (
	"github.com/akiross/gogp/gp"
	"github.com/akiross/gogp/image/draw2d/imgut"
	"github.com/akiross/gogp/node"
	"github.com/akiross/gogp/repr/expr/binary"
	//	"math/rand"
)

/***********************************
 * Genetic Operators
 **********************************/

// Define primitives
var Functionals []gp.Primitive = []gp.Primitive{}
var Terminals []gp.Primitive = []gp.Primitive{}

func Draw(ind *node.Node, img *imgut.Image) {
	// We have to compile the nodes
	exec := node.CompileTree(ind).(*binary.Primitive)
	// Apply the function
	//	exec(0 0, float64(img.W), float64(img.H), img)
	var call imgut.PixelFunc = func(x, y float64) float64 {
		return float64(exec.Eval(binary.NumericIn(x), binary.NumericIn(y)))
	}
	img.FillMathBounds(call)
}

func MakeMaxDepth(limit int) func(*imgut.Image) int {
	return func(img *imgut.Image) int {
		return limit
	}
}

func init() {
	// Add some random constants to terminals
	//	for i := 0; i < 20; i++ {
	//		c := binary.NumericOut(rand.Float64()*100.0 - 50.0)
	//		Terminals = append(Terminals, binary.Terminal(binary.Constant(c)))
	//	}
}
