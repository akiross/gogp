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
	"container/list"
	"fmt"
	"github.com/akiross/gogp/gp"
	"math"
	"math/rand"
)

// strategy is a function that pick a gp.Primitive suitable to be placed in the tree, indicating if it's a functional or not, and giving the index of the picked gp.Primitive
func makeTree(depth int, funcs, terms []gp.Primitive, strategy func(int, int, int) (bool, int)) (root *Node) {
	nFuncs, nTerms := len(funcs), len(terms)
	nType, k := strategy(depth, nFuncs, nTerms)

	if nType {
		// Functional
		root = &Node{funcs[k], nil}
		root.children = make([]*Node, funcs[k].Arity())
		for i := range root.children {
			root.children[i] = makeTree(depth-1, funcs, terms, strategy)
		}
	} else {
		if terms[k].IsEphemeral() {
			root = &Node{terms[k].Run(), nil}
		} else {
			root = &Node{terms[k], nil}
		}
	}
	return // No need to go down for terminals
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
		if root.value.IsEphemeral() {
			panic("Ephemerals should never be in a compilable tree!")
		}
		return root.value.Run()
	}
}

// Mutate the tree by changing one single node with an equivalent one in arity
// funcs is the set of functionals (internal nodes)
// terms is the set of terminals (leaves)
func MakeTreeSingleMutation(funcs, terms []gp.Primitive) func(*Node) {
	return func(t *Node) {
		// Get a slice with the nodes
		nodes, _, _ := t.Enumerate()
		size := len(nodes)
		// Pick a random node
		nid := rand.Intn(size)
		// We can just change the gp.Primitive for that node, picking a similar one
		arity := nodes[nid].value.Arity()
		if arity <= 0 {
			// Terminals have non-positive arity
			k := rand.Intn(len(terms))
			if terms[k].IsEphemeral() {
				nodes[nid].value = terms[k].Run()
			} else {
				nodes[nid].value = terms[k]
			}
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
func MakeTreeNodeMutation(funcs, terms []gp.Primitive) func(float64, *Node) int {
	// Build a map of primitives by arity
	prims := make(map[int][]gp.Primitive)
	for i := range funcs {
		arity := funcs[i].Arity()
		if _, ok := prims[arity]; ok {
			prims[arity] = append(prims[arity], funcs[i])
		} else {
			prims[arity] = make([]gp.Primitive, 1)
			prims[arity][0] = funcs[i]
		}
	}
	// Terminals have arity -1
	prims[-1] = terms

	return func(pMut float64, t *Node) int {
		// Get a slice with the nodes
		nodes, _, _ := t.Enumerate()
		mutCount := 0
		for i := range nodes {
			// For each node, check if it should be mutated
			if rand.Float64() < pMut {
				continue
			}
			// If mutation occurs, pick a random node of same arity
			arity := nodes[i].value.Arity()
			nid := rand.Intn(len(prims[arity]))
			if prims[arity][nid].IsEphemeral() {
				nodes[i].value = prims[arity][nid].Run()
			} else {
				nodes[i].value = prims[arity][nid]
			}
			mutCount += 1
		}
		return mutCount
	}
}

// Swap node contents
func swapNodes(n1, n2 *Node) {
	n1.value, n2.value = n2.value, n1.value
	n1.children, n2.children = n2.children, n1.children
}

// Randomly select two subrees and swap them
func MakeSubtreeSwapMutation(funcs, terms []gp.Primitive) func(*Node) {
	return func(t *Node) {
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

// Given a tree (t*) and a node in it (nid), generates a random tree with the
// appropriate height and replace it with the node
func generateHLimitedAndSwap(tNodes []*Node, tDepths, tHeights []int, maxH, nid int, genFunction func(maxH int) *Node) {
	// The random tree cannot make tree larger
	hLimit := maxH - tDepths[nid] - tHeights[nid]
	// Build the replacement
	replacement := genFunction(hLimit)
	// Swap the content of the nodes
	swapNodes(tNodes[nid], replacement)
}

// Replaces a randomly selected subtree with another randomly created subtree
// maxH describes the maximum height of the resulting tree
func MakeSubtreeMutation(maxH int, genFunction func(maxH int) *Node) func(*Node) {
	return func(t *Node) {
		// Get a slice with the nodes
		tNodes, tDepths, tHeights := t.Enumerate()
		size := len(tNodes)
		// Pick a random node
		nid := rand.Intn(size)
		generateHLimitedAndSwap(tNodes, tDepths, tHeights, maxH, nid, genFunction)
	}
}

func partitionLeaves(tNodes []*Node) (leaves, nonleaves []int) {
	size := len(tNodes)
	// Prepare a slice for internal nodes and one for terminals
	leaves = make([]int, 0, size)
	nonleaves = make([]int, 0, size)
	for i := range tNodes {
		if len(tNodes[i].children) == 0 {
			leaves = append(leaves, i)
		} else {
			nonleaves = append(nonleaves, i)
		}
	}
	return
}

func makeExpProbs(tDepths, leaves []int, exp float64) (probs []float64, index []int) {
	// Compute probabilities for each leaf using its depth
	probs = make([]float64, len(leaves))
	index = make([]int, len(leaves))
	for i := range probs {
		probs[i] = math.Pow(exp, float64(-tDepths[leaves[i]]))
		index[i] = i
	}
	return
}

// same as MakeSubtreeMutation, but probability of picking nodes varies:
// 50% of the times, internal nodes (functionals) are mutated with uniform probability
// 50% of the times, leave nodes (terminals) are mutated with probability 1/exp^depth
func MakeSubtreeMutationLevelExp(maxH int, exp float64, genFunction func(maxH int) *Node) func(*Node) {
	return func(t *Node) {
		// Enumerate the nodes
		tNodes, tDepths, tHeights := t.Enumerate()
		leaves, nonleaves := partitionLeaves(tNodes)
		var nid int
		// Pick one category
		if rand.Intn(2) == 0 {
			// Pick randomly one non-leaf and work like subtree mutation
			nid = nonleaves[rand.Intn(len(nonleaves))]
		} else {
			// Pick leaves using non-uniform probability
			probs, index := makeExpProbs(tDepths, leaves, exp)
			// Get a random element according to probabilities
			computeCDFinPlace(probs, index)
			e := extractCFDinPlace(probs)
			nid = index[e]
		}
		generateHLimitedAndSwap(tNodes, tDepths, tHeights, maxH, nid, genFunction)
	}
}

// ProbComputer should return a pure function that associates to each node of the input
// tree a likelihood of being randomly selected. The likelihoods will be
// normalized before computing the CDF
// The function is expected to be pure, i.e. free of side effects
type ProbComputer func(t *Node) func(*Node) float64

// Subtree mutation, but uses an external tree to determine node mutation probabilities
func MakeSubtreeMutationGuided(maxH int, genFunction func(maxH int) *Node, pc ProbComputer) func(*Node) {
	return func(t *Node) {
		nl := pc(t)                                // Compute nodes likelihood function
		tNodes, tDepths, tHeights := t.Enumerate() // Enumerate nodes
		probs := make([]float64, len(tNodes))
		inds := make([]int, len(tNodes))
		for i, v := range tNodes {
			inds[i] = i
			probs[i] = nl(v)
		}
		normalSlice(probs)              // Normalize likelihood
		computeCDFinPlace(probs, inds)  // Compute CDF slice
		nid := extractCFDinPlace(probs) // Extract node index
		// Perform the mutation
		generateHLimitedAndSwap(tNodes, tDepths, tHeights, maxH, inds[nid], genFunction)
	}
}

// Assign to the nodes a probability of being selected which is
func ArityDepthProbComputer(t *Node) func(*Node) float64 {
	// Enumerate the tree and get node parents and depths
	curDep, depths := 0, make(map[Tree]int) // Depth of each node

	parents := make(map[Tree]Tree) // Associate to each node its parent
	parStack := list.New()         // Stack of nodes being explored
	parStack.PushBack(t)           // Start with the root as parent of itself

	nLeafs := 0 // How many leafs in the tree

	ent := func(n Tree) {
		depths[n] = curDep
		curDep++

		p := parStack.Back().Value.(Tree) // Get current parent
		parents[n] = p                    // Save parent of current node
		parStack.PushBack(n)              // Set current node as parent

		// Increase leaves count if necessary
		if n.ChiCo() == 0 {
			nLeafs++
		}
	}
	exi := func(n Tree) {
		curDep--

		parStack.Remove(parStack.Back()) // Pop current node from stack
	}
	Traverse(t, ent, exi)

	// Likelihood of internal nodes
	inl := 1.0 / float64(len(depths)-nLeafs)

	return func(r *Node) float64 {
		if r.ChiCo() != 0 {
			return float64(inl)
		} else {
			// Leafs have a probability proportional to
			// their depth and to parent's arity
			l := 1
			n := Tree(r)
			for n != parents[n] {
				l = l * parents[n].ChiCo()
				n = parents[n]
			}
			return 1.0 / float64(l)
		}
	}
}

// Counts the nodes in the tree and assign to each node the same probability
func UniformProbComputer(t *Node) func(*Node) float64 {
	return func(r *Node) float64 {
		return 1 // Same likelihood for every node
	}
}

/*
// Applies multiple mutations at random
func MakeMultiMutation(maxH int, genFunction func(maxH int) *Node, funcs, terms []gp.Primitive) func(float64, *Node) bool {
	mutFuncs := make([]func(float64, *Node) bool, 3)
	mutFuncs[0] = MakeTreeSingleMutation(funcs, terms)
	mutFuncs[1] = MakeTreeNodeMutation(funcs, terms)
	mutFuncs[2] = MakeSubtreeMutation(maxH, genFunction)
	//	mutFuncs[3] = MakeLocalNodeMutation(funcs, terms)

	return func(pMut float64, t *Node) bool {
		i := rand.Intn(len(mutFuncs))
		return mutFuncs[i](pMut, t)
	}
}
*/

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
func MakeTree1pCrossover(maxDepth int) func(_, _ *Node) {
	return func(t1, t2 *Node) {
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
