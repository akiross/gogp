package gogp

import (
	"ale-re.net/phd/imgut"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strings"
)

// The function that will be called to get a solution
type Terminal func(x1, x2, y1, y2 float64, img *imgut.Image)

// Functionals are internal nodes of a solution tree
type Functional func(args ...Terminal) Terminal

// Buils a Terminal that fills the entire area with given color
func RectFiller(col ...float64) Terminal {
	return func(x1, y1, x2, y2 float64, img *imgut.Image) {
		img.FillRect(x1, y1, x2, y2, col...)
	}
}

// Returns a terminal that fills the rectangle according to left and right
func VSplit(args ...Terminal) Terminal {
	return func(x1, y1, x2, y2 float64, img *imgut.Image) {
		xh := (x1 + x2) * 0.5
		args[0](x1, y1, xh, y2, img)
		args[1](xh, y1, x2, y2, img)
	}
}

func HSplit(args ...Terminal) Terminal {
	return func(x1, y1, x2, y2 float64, img *imgut.Image) {
		yh := (y1 + y2) * 0.5
		args[0](x1, y1, x2, yh, img)
		args[1](x1, yh, x2, y2, img)
	}
}

type TriTerminal func(x1, x2, y float64, img *imgut.Image)

type TriFunctional func(top, left, center, right TriTerminal) TriTerminal

// Return a terminal that fills the entire triangle with given color
func TriFiller(col ...float64) TriTerminal {
	return func(x1, x2, y float64, img *imgut.Image) {
		img.FillTriangle(x1, x2, y, col...)
	}
}

func TriSplit(top, left, center, right TriTerminal) TriTerminal {
	return func(x1, x2, y float64, img *imgut.Image) {
		// Split the triangle in 4 parts
		cx1, cxm, cx2 := x1+0.25*(x2-x1), x1+0.5*(x2-x1), x1+0.75*(x2-x1)
		ty := imgut.TriangleCenterY(x1, x2, y)
		cy := 0.5 * (ty + y)

		top(cx1, cx2, cy, img)
		left(x1, cxm, y, img)
		center(cx2, cx1, cy, img)
		right(cxm, x2, y, img)
	}
}

type Node struct {
	value    interface{} // Any type
	children []*Node
}

func funcName(f interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	s := strings.Split(name, ".")
	return s[len(s)-1]
}

func (root *Node) String() string {
	if len(root.children) == 0 {
		// Terminal node
		return fmt.Sprintf("T{%v}", funcName(root.value))
	} else {
		sChildren := root.children[0].String()
		for i := 1; i < len(root.children); i++ {
			sChildren += ", " + root.children[i].String()
		}
		return fmt.Sprintf("F{%v}(%v)", funcName(root.value), sChildren)
	}
}

func makeTree(depth, arity int, funcs []Functional, terms []Terminal, strategy func(int, int, int) (bool, int)) (root *Node) {
	nFuncs, nTerms := len(funcs), len(terms)
	nType, k := strategy(depth, nFuncs, nTerms)

	if nType {
		root = &Node{funcs[k], nil}
		root.children = make([]*Node, arity)
		for i := range root.children {
			root.children[i] = makeTree(depth-1, arity, funcs, terms, strategy)
		}
		return
	} else {
		root = &Node{terms[k], nil}
		return // No need to go down for terminals
	}
}

func MakeTreeGrow(maxH, arity int, funcs []Functional, terms []Terminal) *Node {
	growStrategy := func(depth, nFuncs, nTerms int) (isFunc bool, k int) {
		if depth == 0 {
			return false, rand.Intn(nTerms)
		} else {
			k := rand.Intn(nFuncs + nTerms)
			if k < nFuncs {
				return true, k
			} else {
				return false, k - nFuncs
			}
		}
	}

	return makeTree(maxH, arity, funcs, terms, growStrategy)
}

func MakeTreeFull(maxH, arity int, funcs []Functional, terms []Terminal) *Node {
	fullStrategy := func(depth, nFuncs, nTerms int) (isFunc bool, k int) {
		if depth == 0 {
			return false, rand.Intn(nTerms)
		} else {
			return true, rand.Intn(nFuncs)
		}
	}
	return makeTree(maxH, arity, funcs, terms, fullStrategy)
}

func CompileTree(root *Node) Terminal {
	switch root.value.(type) {
	case Terminal:
		return root.value.(Terminal)
	case Functional:
		// If it's a functional, compile each children and return
		terms := make([]Terminal, len(root.children))
		for i := 0; i < len(root.children); i++ {
			terms[i] = CompileTree(root.children[i])
		}
		return (root.value.(Functional))(terms...)
	default:
		return nil
	}
}
