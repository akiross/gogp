/*
Package gogp provides the foundation to use Genetic Programming (GP) in Go.

The package provides basic functions and methods to implement a GP system, such as:

- Node mutation

- One point crossover

The code is organized around trees, therefore only tree-like structures are admitted.

To use gogp, you have to create the gp.Primitives (functionals and terminals) that will
compose your solutions.

1. Create a type for terminals: this type usually identifies the number of variables
that you are using as input (i.e. how many input variables do you have?)

2. Create a type for every functional group that you are about to use: each group is
characterized by the same number of parameters (arity). For example, if you use
unary and binary functionals, you need two types.

3. Implement your operators.

*/
package node

import (
	"fmt"
	"github.com/akiross/gogp/gp"
	"math/rand"
)

// strategy is a function that pick a gp.Primitive suitable to be placed in the tree, indicating if it's a functional or not, and giving the index of the picked gp.Primitive
func makeTree(depth int, funcs, terms []gp.Primitive, strategy func(int, int, int) (bool, int)) (root *Node) {
	nFuncs, nTerms := len(funcs), len(terms)
	nType, k := strategy(depth, nFuncs, nTerms)

	if nType {
		root = &Node{funcs[k], nil}
		root.children = make([]*Node, funcs[k].Arity())
		for i := range root.children {
			root.children[i] = makeTree(depth-1, funcs, terms, strategy)
		}
		return
	} else {
		root = &Node{terms[k], nil}
		return // No need to go down for terminals
	}
}

// Builds a tree using the grow method
func MakeTreeGrow(maxH int, funcs, terms []gp.Primitive) *Node {
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
	return makeTree(maxH, funcs, terms, growStrategy)
}

// Builds a tree using the grow method, but pick 50-50 funcs and terms
func MakeTreeGrowBalanced(maxH int, funcs, terms []gp.Primitive) *Node {
	growBalStrategy := func(depth, nFuncs, nTerms int) (isFunc bool, k int) {
		if depth == 0 {
			return false, rand.Intn(nTerms)
		} else {
			if rand.Intn(2) == 0 {
				// Pick a random functional
				return true, rand.Intn(nFuncs)
			} else {
				// Pick a random terminal
				return false, rand.Intn(nTerms)
			}
		}
	}
	return makeTree(maxH, funcs, terms, growBalStrategy)
}

// Builds a tree using the full method
func MakeTreeFull(maxH int, funcs, terms []gp.Primitive) *Node {
	fullStrategy := func(depth, nFuncs, nTerms int) (isFunc bool, k int) {
		if depth == 0 {
			return false, rand.Intn(nTerms)
		} else {
			return true, rand.Intn(nFuncs)
		}
	}
	return makeTree(maxH, funcs, terms, fullStrategy)
}

// Make a tree using the half and half method.
// It's not the ramped version: it just uses grow or full with 50% chances
func MakeTreeHalfAndHalf(maxH int, funcs, terms []gp.Primitive) *Node {
	if rand.Intn(2) == 0 {
		return MakeTreeGrowBalanced(maxH, funcs, terms)
	} else {
		return MakeTreeFull(maxH, funcs, terms)
	}
}

// Compiles a tree returning a gp.Primitive, resulting
// from the execution of the Run method
func CompileTree(root *Node) gp.Primitive {
	if root.value.IsFunctional() {
		// If it's a functional, compile each children and return
		terms := make([]gp.Primitive, len(root.children))
		for i := 0; i < len(root.children); i++ {
			terms[i] = CompileTree(root.children[i])
		}
		if root.value.Arity() != len(terms) {
			fmt.Println("ERROR! Trying to call a Functional with Arity", root.value.Arity(), "passing", len(terms), "arguments")
			fmt.Println("Tree is", root)
			panic(fmt.Sprint("ERROR! Trying to call a Functional with Arity ", root.value.Arity(), " passing ", len(terms), " arguments: ", gp.FuncName(root.value)))
		}
		return root.value.Run(terms...)
	} else {
		return root.value.Run()
	}
}

// Mutate the tree by changing one single node with an equivalent one in arity
// funcs is the set of functionals (internal nodes)
// terms is the set of terminals (leaves)
func MakeTreeSingleMutation(funcs, terms []gp.Primitive) func(float64, *Node) {
	return func(pMut float64, t1 *Node) {
		// Check if it's necessary to mutate the node
		if rand.Float64() >= pMut {
			return
		}
		// Get a slice with the nodes
		nodes, _, _ := t1.Enumerate()
		size := len(nodes)
		// Pick a random node
		nid := rand.Intn(size)
		// We can just change the gp.Primitive for that node, picking a similar one
		arity := nodes[nid].value.Arity()
		if arity <= 0 {
			// Terminals have non-positive arity
			nodes[nid].value = terms[rand.Intn(len(terms))]
		} else {
			// Functionals
			sameArityFuncs := make([]gp.Primitive, 0, len(funcs))
			for i := 0; i < len(funcs); i++ {
				if funcs[i].Arity() == arity {
					sameArityFuncs = append(sameArityFuncs, funcs[i])
				}
			}
			// Replace node with a random one
			nodes[nid].value = sameArityFuncs[rand.Intn(len(sameArityFuncs))]
		}
	}
}

// Go over each node and randomly mutate it with a compatible one
func MakeTreeNodeMutation(funcs, terms []gp.Primitive) func(float64, *Node) {
	// Build a map of primitives by arity
	prims := make(map[int][]gp.Primitive)
	for i := range funcs {
		arity := funcs[i].Arity()
		if _, ok := prims[arity]; ok {
			prims[i] = append(prims[i], funcs[i])
		} else {
			prims[i] = make([]gp.Primitive, 1)
			prims[i][0] = funcs[i]
		}
	}
	// Terminals have arity -1
	prims[-1] = terms

	return func(pMut float64, t1 *Node) {
		// Get a slice with the nodes
		nodes, _, _ := t1.Enumerate()
		for i := range nodes {
			// For each node, check if it should be mutated
			if rand.Float64() < pMut {
				continue
			}
			// If mutation occurs, pick a random node of same arity
			arity := nodes[i].value.Arity()
			nid := rand.Intn(len(prims[arity]))
			nodes[i].value = prims[arity][nid]
		}
	}
}

// Swap node contents
func swapNodes(n1, n2 *Node) {
	n1.value, n2.value = n2.value, n1.value
	n1.children, n2.children = n2.children, n1.children
}

// Randomly select two subrees and swap them
func MakeSubtreeSwapMutation(funcs, terms []gp.Primitive) func(float64, *Node) {
	return func(pMut float64, t *Node) {
		// The tricky part is to pick two subtrees that are distinct
		// i.e. we do not want that one tree is subtree of the other
		// How? We can get a list of nodes and pick one, then
		// we remove children and granchildren from the list, and also
		// the parents. We then pick another tree from nodes left

		// Get the list of nodes
		//nodes, _, _ := t.Enumerate()
		// Get a random node
		//n1 := rand.Intn(len(nodes))
		// Enumerate unpickable nodes
		//unpickable, _, _ := nodes[n1].Enumerate()
		// Trovare i genitori non è veloce... Non c'é puntatore al parent!
		panic("NOT IMPLEMENTED YET")
	}
}

// Replaces a randomly selected subtree with another randomly created subtree
// maxH describes the maximum height of the resulting tree
func MakeSubtreeMutation(maxH int, genFunction func(maxH int) *Node) func(float64, *Node) {
	return func(pMut float64, t *Node) {
		// Check if mutation is necessary
		if rand.Float64() >= pMut {
			return
		}

		// Get a slice with the nodes
		tNodes, tDepths, tHeights := t.Enumerate()
		size := len(tNodes)
		// Pick a random node
		nid := rand.Intn(size)
		// The random tree cannot make tree larger
		hLimit := maxH - tDepths[nid] - tHeights[nid]
		// Build the replacement
		replacement := genFunction(hLimit)
		// Swap the content of the nodes
		swapNodes(tNodes[nid], replacement)
		// Replacement is discarded
	}
}

/*
func intMax(n ...int) int {
	index := 0
	for i := 1; i < len(n); i++ {
		if n[i] > n[index] {
			index = i
		}
	}
	return n[index]
}
*/

// Height-limited crossover, to prevent bloating
func MakeTree1pCrossover(maxDepth int) func(float64, *Node, *Node) {
	return func(pCross float64, t1, t2 *Node) {
		// Check if crossover is necessary
		if rand.Float64() >= pCross {
			return
		}

		// Get the slices for the trees, including node heights
		t1Nodes, t1Depths, t1Heights := t1.Enumerate()
		t2Nodes, t2Depths, t2Heights := t2.Enumerate()

		// Max depth of each tree is given by the max distance
		// of the root from the leaves
		t1MaxDe, t2MaxDe := t1Heights[0], t2Heights[0]

		// If the max depth is too small, it's an error, panic!
		if t1MaxDe > maxDepth || t2MaxDe > maxDepth {
			panic(fmt.Sprintf("ERROR! maxHeight (%v) is lower than tree depth(s) (%v and %v)", maxDepth, t1MaxDe, t2MaxDe))
		}

		// Indices for the random nodes to swap, in t1Nodes and t2Nodes
		var rn1, rn2 int
		if maxDepth < 0 {
			// No bloat control, pick two random nodes
			rn1, rn2 = rand.Intn(len(t1Nodes)), rand.Intn(len(t2Nodes))
		} else {
			// Bloat control: pick one first, then the other
			rn1 = rand.Intn(len(t1Nodes))
			// Copy only the index of nodes allowed to be picked
			// A node n2 in t2 can be picked after the picking of node n1 in t1 if:
			//   depth(n1) + height(n2) <= MaxDepth AND depth(n2) + height(n1) <= MaxDepth
			allowed := make([]int, 0, len(t2Nodes))
			for i := 0; i < len(t2Nodes); i++ {
				if (t1Depths[rn1]+t2Heights[i] <= maxDepth) && (t2Depths[i]+t1Heights[rn1] <= maxDepth) {
					//				if t2Heights[i] <= maxHeight {
					allowed = append(allowed, i)
				}
			}
			if len(allowed) == 0 {
				panic("This should not be empty!!")
			}
			// Take a node in the allowed set
			rn2 = allowed[rand.Intn(len(allowed))]
		}

		// Swap the nodes in the parent, so that references to nodes will be valid
		// Swap the content of the nodes (so, we can swap also roots)
		swapNodes(t1Nodes[rn1], t2Nodes[rn2])
	}
}