package expr

import (
	"github.com/akiross/gogp/gp"
	"github.com/akiross/gogp/image/draw2d/imgut"
	"github.com/akiross/gogp/node"
	"github.com/akiross/gogp/repr/expr/binary"
	"math/rand"
)

/***********************************
 * Genetic Operators
 **********************************/

// Define primitives
var Functionals []gp.Primitive = []gp.Primitive{
	binary.Functional3(binary.Choice), // if-then-else
	binary.Functional2(binary.Sum),
	binary.Functional2(binary.Sub),
	binary.Functional2(binary.Mul),
	binary.Functional2(binary.ProtectedDiv),
	binary.Functional2(binary.Min),
	binary.Functional2(binary.Max),
	binary.Functional2(binary.Pow),
	binary.Functional1(binary.Sqrt),
	binary.Functional1(binary.Square),
	binary.Functional1(binary.Abs),
	binary.Functional1(binary.Neg),
	binary.Functional1(binary.Sign),
}

var Terminals []gp.Primitive = []gp.Primitive{
	binary.Terminal(binary.IdentityX),
	binary.Terminal(binary.IdentityY),
	binary.Terminal(binary.Constant(-1)),
	binary.Terminal(binary.Constant(0)),
	binary.Terminal(binary.Constant(1)),
	binary.Terminal(binary.Constant(2)),
	binary.Terminal(binary.Constant(10)),
}

func Draw(ind *node.Node, img *imgut.Image) {
	// We have to compile the nodes
	exec := node.CompileTree(ind).(binary.Terminal)
	// Apply the function
	//	exec(0 0, float64(img.W), float64(img.H), img)
	var call imgut.PixelFunc = func(x, y float64) float64 {
		return float64(exec(binary.NumericIn(x), binary.NumericIn(y)))
	}
	img.FillMathBounds(call)
}

func MaxDepth(img *imgut.Image) int {
	// Depth is fixed, we cannot get a "feature size" depending on image size
	return 10
}

func init() {
	// Add some random constants to terminals
	for i := 0; i < 20; i++ {
		c := binary.NumericOut(rand.Float64()*100.0 - 50.0)
		Terminals = append(Terminals, binary.Terminal(binary.Constant(c)))
	}
}
