package gogp

import (
	"bytes"
	"container/list"
	"fmt"
	"math/rand"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
)

type Primitive interface {
	IsFunctional() bool         // Returns true if is functional, false if terminal
	Arity() int                 // Returns the arity of the primitive
	Run(...Primitive) Primitive // Functionals returns terminals, terminals do nothing
}

type Node struct {
	value    Primitive
	children []*Node
}

func funcName(f Primitive) string {
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

// Enumerate the path of the nodes, where the path is the
// number of the node in parent's children up to the root
// e.g. leftmost node in a binary full tree with depth 4
// is 0000, rightmost is 1111. Root has empty path
func (root *Node) Path() map[*Node][]int {
	// This map stores the path for every node in the tree
	paths := make(map[*Node][]int)

	// This visit function will save the values in the paths map
	var bfVisit func(*Node, []int)
	bfVisit = func(r *Node, curPath []int) {
		// Store the current path in the node
		paths[r] = curPath
		// For each child, build the path
		for i := range r.children {
			// Copy the current path
			cPath := append([]int{}, curPath...)
			// Append the number of children
			cPath = append(cPath, i)
			// Go down
			bfVisit(r.children[i], cPath)
		}
	}

	bfVisit(root, []int{})

	return paths
}

// Get the nodes in breadth-first order
func (root *Node) BreadthFirstEnum() []*Node {
	// Queue of nodes to visit
	queue := list.New()
	queue.PushBack(root)
	// List of visited nodes
	visited := []*Node{root}

	for queue.Front() != nil {
		n := queue.Front()
		visited = append(visited, n.Value.(*Node))
		for _, c := range n.Value.(*Node).children {
			queue.PushBack(c)
		}
		queue.Remove(n)
	}

	return visited
}

// Colorize a tree using a stating hue value, with varying deviation
func (root *Node) Colorize(hue, hDev float32) map[*Node][]float32 {
	// Queue of nodes to visit
	queue := list.New()
	queue.PushBack(root)
	// Queue for depths
	depths := list.New()
	depths.PushBack(0)
	// List of visited nodes
	visited := []*Node{root}
	// Colors of the visited nodes
	colors := make(map[*Node][]float32)

	// Compute the variation of saturation for each node depth
	treeDepth := root.Depth()
	var satVar float32 = 0.9 / float32(treeDepth)

	for queue.Front() != nil {
		ne := queue.Front()   // Node element
		n := ne.Value.(*Node) // Node
		de := depths.Front()  // Depth element
		d := de.Value.(int)   // node depth
		visited = append(visited, n)
		for _, c := range n.children {
			queue.PushBack(c)
			depths.PushBack(d + 1)
		}

		// Set the HSV color
		colors[n] = []float32{hue, 1.0 - float32(d)*satVar, 0.8}

		queue.Remove(ne)
		depths.Remove(de)
	}

	return colors
}

func (root *Node) PrettyPrint() string {
	var ind func(r *Node, d int, tabStr string) string
	ind = func(r *Node, d int, tabStr string) string {
		str := ""
		for i := 0; i < d; i++ {
			str += tabStr
		}
		str += fmt.Sprintf("%v\n", funcName(r.value))
		for i := 0; i < len(r.children); i++ {
			str += ind(r.children[i], d+1, tabStr)
		}
		return str
	}
	return ind(root, 0, "    ")
}

func (root *Node) GraphvizDot(name string, hsvColors map[*Node][]float32) string {
	// Function to build edges
	var ind func(parent, r *Node, tabStr string) string

	// Store the identifiers for tree node
	ids := make(map[*Node]string)
	// Count how many times the name has been used
	counts := make(map[string]int)

	ind = func(parent, r *Node, tabStr string) string {
		// Get the identifier for the node, and see if it's present
		_, ok := ids[r]
		if !ok {
			// This is the first time we use this node, give it a name
			name := funcName(r.value)
			// Increase the sequential numbet for this node type
			counts[name] += 1
			ids[r] = fmt.Sprintf("\"%v-%v\"", name, counts[name])
		}

		//
		str := ""

		if parent != nil {
			// We print the "parent -> children;" line
			str += fmt.Sprintf("%v%v -> %v;\n", tabStr, ids[parent], ids[r])
		}
		for i := 0; i < len(r.children); i++ {
			str += ind(r, r.children[i], tabStr)
		}
		return str
	}

	// Write edges, and also modify the identifiers map
	edges := ind(nil, root, "    ")

	// If a color is present, we set the fill color

	style := ""
	colors := ""
	if hsvColors != nil {
		style = "node [style=filled,margin=0,shape=box];\n"
		for k, v := range ids {
			colors += fmt.Sprintf("%v [fillcolor=\"%0.3f %0.3f %0.3f\", label=\"%p\"]\n", v, hsvColors[k][0], hsvColors[k][1], hsvColors[k][2], k)
		}
	}

	src := fmt.Sprintf("digraph %v {rankdir=LR;%v\n%v\n%v\n label=\"%v\"}\n", name, style, edges, colors, name)

	cmd := exec.Command("dot", "-Tpng", "-o", name+".png")
	cmd.Stdin = strings.NewReader(src)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Cannot run dot (%v)\nOutput:\n%v\nError:\n%v\n", err, out.String(), stderr.String())
		return src
	}
	return "" //src
}

func makeTree(depth, arity int, funcs, terms []Primitive, strategy func(int, int, int) (bool, int)) (root *Node) {
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

func MakeTreeGrow(maxH int, funcs, terms []Primitive) *Node {
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
	arity := funcs[0].Arity()
	return makeTree(maxH, arity, funcs, terms, growStrategy)
}

func MakeTreeFull(maxH int, funcs, terms []Primitive) *Node {
	fullStrategy := func(depth, nFuncs, nTerms int) (isFunc bool, k int) {
		if depth == 0 {
			return false, rand.Intn(nTerms)
		} else {
			return true, rand.Intn(nFuncs)
		}
	}
	arity := funcs[0].Arity()
	return makeTree(maxH, arity, funcs, terms, fullStrategy)
}

// It's not the ramped version: it just uses grow or full with 50% chances
func MakeTreeHalfAndHalf(maxH int, funcs, terms []Primitive) *Node {
	if rand.Intn(2) == 0 {
		return MakeTreeFull(maxH, funcs, terms)
	} else {
		return MakeTreeGrow(maxH, funcs, terms)
	}
}

func CompileTree(root *Node) Primitive {
	if root.value.IsFunctional() {
		// If it's a functional, compile each children and return
		terms := make([]Primitive, len(root.children))
		for i := 0; i < len(root.children); i++ {
			terms[i] = CompileTree(root.children[i])
		}
		return root.value.Run(terms...)
	} else {
		return root.value
	}
}

// Depth is the max number of edges between root and a leaf
func (root *Node) Depth() int {
	if len(root.children) == 0 {
		return 0
	}
	h := root.children[0].Depth()
	for i := 1; i < len(root.children); i++ {
		h2 := root.children[i].Depth()
		if h2 > h {
			h = h2
		}
	}
	return 1 + h
}

// The number of nodes in the tree
func (root *Node) Size() int {
	if len(root.children) == 0 {
		return 1
	}
	c := 1
	for i := 1; i < len(root.children); i++ {
		c += root.children[i].Size()
	}
	return c
}

// Visit the tree and transform it to a slice of nodes, and a slice for their depths and heights
// FIXME this can be (partially) avoided by usind directly slices to store the tree (using arity)
func (root *Node) Enumerate() ([]*Node, []int, []int) {
	tree := make([]*Node, 1)
	dtree := make([]int, 1) // Depth of a node (node-to-root)
	htree := make([]int, 1) // Node heights (node-to-leaf)
	tree[0] = root
	dtree[0] = 0
	var chhe []int // Storage for children heights
	// TODO I think this could be made faster by joining the appends in a single loop
	for i := 0; i < len(root.children); i++ {
		chtr, chde, chhe2 := root.children[i].Enumerate()
		// Add 1 to the depth of the child
		for j := range chde {
			chde[j]++
		}
		// Append chid slices
		chhe = append(chhe, chhe2...)
		tree = append(tree, chtr...)
		dtree = append(dtree, chde...)
	}
	// Compute height of nodes
	if len(root.children) == 0 {
		htree[0] = 0 // Leaves have height 0
	} else {
		// Node's depth is 1+max(chhe)
		max := chhe[0]
		for i := 1; i < len(chhe); i++ {
			if chhe[i] > max {
				max = chhe[i]
			}
		}
		htree[0] = max + 1
	}
	htree = append(htree, chhe...)

	return tree, dtree, htree
}

// Mutate the tree by changing one single node with an equivalent one in arity
// funcs is the set of functionals (internal nodes)
// terms is the set of terminals (leaves)
func MakeTreeNodeMutation(funcs, terms []Primitive) func(*Node) {
	return func(t1 *Node) {
		// Get a slice with the nodes
		nodes, _, _ := t1.Enumerate()
		size := len(nodes)
		// Pick a random node
		nid := rand.Intn(size)
		// We can just change the primitive for that node, picking a similar one
		arity := nodes[nid].value.Arity()
		if arity <= 0 {
			// Terminals have non-positive arity
			nodes[nid].value = terms[rand.Intn(len(terms))]
		} else {
			// Functionals
			sameArityFuncs := make([]Primitive, 0, len(funcs))
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
func MakeTree1pCrossover(maxDepth int) func(*Node, *Node) {
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
		n1, n2 := t1Nodes[rn1], t2Nodes[rn2]

		// Swap the content of the nodes (so, we can swap also roots)
		n1.value, n2.value = n2.value, n1.value
		n1.children, n2.children = n2.children, n1.children
	}
}
