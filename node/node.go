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
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"github.com/akiross/gogp/gp"
	"os/exec"
	"strings"
)

type Node struct {
	value    gp.Primitive
	children []*Node
}

func (root *Node) String() string {
	if len(root.children) == 0 {
		// Terminal node
		return fmt.Sprintf("T{%v}", root.value.Name())
	} else {
		sChildren := root.children[0].String()
		for i := 1; i < len(root.children); i++ {
			sChildren += ", " + root.children[i].String()
		}
		return fmt.Sprintf("F{%v}(%v)", root.value.Name(), sChildren)
	}
}

func (root *Node) MarshalJSON() ([]byte, error) {
	if len(root.children) == 0 {
		return json.Marshal(map[string]interface{}{
			"terminal": root.value.Name(),
		})
	} else {
		return json.Marshal(map[string]interface{}{
			"functional": root.value.Name(),
			"children":   root.children,
		}) //.children) //[]byte(`{"node":"foo"}`), nil
	}
}

// Full copy of the tree
func (root *Node) Copy() *Node {
	childr := make([]*Node, len(root.children))
	for i := range root.children {
		childr[i] = root.children[i].Copy()
	}
	return &Node{root.value, childr}
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

// Prints the tree using a nice indentation
func (root *Node) PrettyPrint() string {
	var ind func(r *Node, d int, tabStr string) string
	ind = func(r *Node, d int, tabStr string) string {
		str := ""
		for i := 0; i < d; i++ {
			str += tabStr
		}
		str += fmt.Sprintf("%v\n", gp.FuncName(r.value))
		for i := 0; i < len(r.children); i++ {
			str += ind(r.children[i], d+1, tabStr)
		}
		return str
	}
	return ind(root, 0, "    ")
}

// Produces a graph for the tree, using graphviz dot, saving the graph
// in the file <name>.png. If a color map (generated with Colorize) is
// provided, the nodes will be colored
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
			name := gp.FuncName(r.value)
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
// BUG(akiross) this can be (partially) avoided by usind directly slices to store the tree (using arity)
func (root *Node) Enumerate() ([]*Node, []int, []int) {
	tree := make([]*Node, 1)
	dtree := make([]int, 1) // Depth of a node (node-to-root)
	htree := make([]int, 1) // Node heights (node-to-leaf)
	tree[0] = root
	dtree[0] = 0
	var chhe []int // Storage for children heights
	// BUG(akiross) I think this could be made faster by joining the appends in a single loop
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
