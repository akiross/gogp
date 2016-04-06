package node

// Generic Tree interface and generic functions
import (
	"container/list"
)

type Tree interface {
	ChiCo() int // Children count
	Child(i int) Tree
}

// Depth is the max number of edges between root and a leaf
func Depth(t Tree) int {
	if t.ChiCo() == 0 {
		return 0
	}
	h := Depth(t.Child(0))
	for i := 1; i < t.ChiCo(); i++ {
		h2 := Depth(t.Child(i))
		if h2 > h {
			h = h2
		}
	}
	return 1 + h
}

// The number of nodes in the tree
func Size(t Tree) int {
	if t.ChiCo() == 0 {
		return 1
	}
	c := 1
	for i := 1; i < t.ChiCo(); i++ {
		c += Size(t.Child(i))
	}
	return c
}

// Enumerate the path of the nodes, where the path is the
// number of the node in parent's children up to the root t
// e.g. leftmost node in a binary full tree with depth 4
// is 0000, rightmost is 1111. Root has empty path
func Path(t Tree) map[Tree][]int {
	// This map stores the path for every node in the tree
	paths := make(map[Tree][]int)

	// This visit function will save the values in the paths map
	var bfVisit func(Tree, []int)
	bfVisit = func(r Tree, curPath []int) {
		// Store the current path in the node
		paths[r] = curPath
		// For each child, build the path
		for i := 0; i < r.ChiCo(); i++ {
			// Copy the current path
			cPath := append([]int{}, curPath...)
			// Append the number of children
			cPath = append(cPath, i)
			// Go down
			bfVisit(r.Child(i), cPath)
		}
	}

	bfVisit(t, []int{})

	return paths
}

// Get the nodes in breadth-first order
func BreadthFirstEnum(t Tree) []Tree {
	// Queue of nodes to visit
	queue := list.New()
	queue.PushBack(t)
	// List of visited nodes
	visited := []Tree{t}

	for queue.Front() != nil {
		n := queue.Front()
		val := n.Value.(Tree)
		visited = append(visited, val)
		for i := 0; i < val.ChiCo(); i++ {
			queue.PushBack(val.Child(i))
		}
		queue.Remove(n)
	}

	return visited
}

// Colorize a tree using a stating hue value, with varying deviation
func Colorize(t Tree, hue, hDev float32) map[Tree][]float32 {
	// Queue of nodes to visit
	queue := list.New()
	queue.PushBack(t)
	// Queue for depths
	depths := list.New()
	depths.PushBack(0)
	// List of visited nodes
	visited := []Tree{t}
	// Colors of the visited nodes
	colors := make(map[Tree][]float32)

	// Compute the variation of saturation for each node depth
	treeDepth := Depth(t)
	var satVar float32 = 0.9 / float32(treeDepth)

	for queue.Front() != nil {
		ne := queue.Front()  // Node element
		n := ne.Value.(Tree) // Node
		de := depths.Front() // Depth element
		d := de.Value.(int)  // node depth
		visited = append(visited, n)
		for i := 0; i < n.ChiCo(); i++ {
			queue.PushBack(n.Child(i))
			depths.PushBack(d + 1)
		}

		// Set the HSV color
		colors[n] = []float32{hue, 1.0 - float32(d)*satVar, 0.8}

		queue.Remove(ne)
		depths.Remove(de)
	}

	return colors
}

// A generic traverse function that allows to specify functions to call before
// each node is visited and after it, and its children, has been visited
func Traverse(t Tree, opEnter, opExit func(Tree)) []Tree {
	tree := []Tree{t}
	if opEnter != nil {
		opEnter(t)
	}
	for i := 0; i < t.ChiCo(); i++ {
		chtr := Traverse(t.Child(i), opEnter, opExit)
		tree = append(tree, chtr...)
	}
	if opExit != nil {
		opExit(t)
	}
	return tree
}

// Enumerate using Traverse
func Enumerate_(t Tree) ([]Tree, []int, []int) {
	curDep, depths := 0, make([]int, 0)
	items := make([]Tree, 0)
	hmap := make(map[Tree]int)
	onEnter := func(r Tree) {
		depths = append(depths, curDep) // Save current depth
		curDep++                        // Increment current depth for children
		items = append(items, r)        // Save item in the list
	}
	onExit := func(r Tree) {
		curDep-- // Children visited, reduce depth
		// Node height is the maximum height of children + 1, if any
		if r.ChiCo() > 0 {
			max := 0
			for i := 0; i < r.ChiCo(); i++ {
				c := r.Child(i)
				if hmap[c] > max {
					max = hmap[c]
				}
			}
			hmap[r] = max + 1
		} else {
			hmap[r] = 0
		}
	}

	Traverse(t, onEnter, onExit)

	heights := make([]int, len(items))
	for i, n := range items {
		heights[i] = hmap[n]
	}
	return items, depths, heights
}

func Enumerate(t Tree) ([]Tree, []int, []int) {
	tree := []Tree{t}
	dtree := []int{0}       // Depth of a node (node-to-root)
	htree := make([]int, 1) // Node heights (node-to-leaf)
	var chhe []int          // Storage for children heights
	// BUG(akiross) I think this could be made faster by joining the appends in a single loop
	for i := 0; i < t.ChiCo(); i++ {
		chtr, chde, chhe2 := Enumerate(t.Child(i))
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
	if t.ChiCo() == 0 {
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
