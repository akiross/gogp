package node

import (
	"container/list"
	"fmt"
	"reflect"
	"testing"
)

type nod struct {
	v        int
	children []*nod
}

func (n *nod) ChiCo() int {
	if n == nil {
		return 0
	}
	return len(n.children)
}

func (n *nod) Child(i int) Tree {
	return n.children[i]
}

func mkn(v int, l, r *nod) *nod {
	return &nod{v, []*nod{l, r}}
}

func mkt(v int) *nod {
	return &nod{v, nil}
}

func (n *nod) String() string {
	if n.ChiCo() == 0 {
		// Terminal node
		return fmt.Sprintf("T{%v}", n.v)
	} else {
		sChildren := fmt.Sprint(n.Child(0))
		for i := 1; i < n.ChiCo(); i++ {
			sChildren += ", " + fmt.Sprint(n.Child(i))
		}
		return fmt.Sprintf("F{%v}(%v)", n.v, sChildren)
	}
}

func skewness(r Tree) float64 {
	if r.ChiCo() == 0 {
		return 0
	} else {
		sk := 0.0
		for i := 0; i < r.ChiCo(); i++ {
			var d, s, j float64
			d = float64(Depth(r.Child(i)))
			s = float64(Size(r.Child(i)))
			if r.ChiCo() > 1 {
				j = 2*float64(i)/float64(r.ChiCo()-1) - 1
			} else {
				j = 0
			}
			sk += j * d * s
		}
		return sk
	}
}

// Measuring the skewness of a tree
func TestSkewness(t *testing.T) {
	tests := []struct {
		expSkew float64
		tree    *nod
	}{
		//		{0.0, mkn(0, nil, nil)},
		//		{-1.0, mkn(0, mkn(1, nil, nil), nil)},
		//		{1.0, mkn(0, nil, mkn(1, nil, nil))},
		{0, mkt(0)},
		{0, mkn(0, mkt(1), mkt(2))},
		{0, mkn(0, mkt(1), mkn(2, mkt(3), mkt(4)))},
		{0, mkn(0, mkn(1, mkt(2), mkt(3)), mkt(4))},
		{0, mkn(0, mkn(1, mkt(2), mkt(3)), mkn(4, mkt(5), mkt(6)))},
		{0, mkn(0, mkn(1, mkn(2, mkt(3), mkt(4)), mkt(5)), mkt(6))},
		{0, mkn(0, mkn(1, mkt(2), mkn(3, mkt(4), mkt(5))), mkt(6))},

		//		{0.0, mkn(0, mkt(1), mkt(2))},
	}

	for _, te := range tests {
		t.Error("Test\n", PrettyPrint(te.tree, func(n Tree) string {
			if n == nil || reflect.ValueOf(n).IsNil() {
				return "[Nil]"
			}
			return fmt.Sprint(n.(*nod).v)
		}), "got skew", skewness(te.tree))
		continue
		if sk := skewness(te.tree); sk != te.expSkew {
			t.Error("Unexpected skewness: got", sk, "expected", te.expSkew)
		} else {
			fmt.Println("Test passed:", te)
		}
	}
}

func TestSize(t *testing.T) {
	tests := []struct {
		size int
		tree *nod
	}{
		{1, mkt(0)},
		{3, mkn(0, mkt(1), mkt(2))},
		{5, mkn(0, mkt(1), mkn(2, mkt(3), mkt(4)))},
		{5, mkn(0, mkn(1, mkt(2), mkt(3)), mkt(4))},
		{7, mkn(0, mkn(1, mkt(2), mkt(3)), mkn(4, mkt(5), mkt(6)))},
		{7, mkn(0, mkn(1, mkn(2, mkt(3), mkt(4)), mkt(5)), mkt(6))},
		{7, mkn(0, mkn(1, mkt(2), mkn(3, mkt(4), mkt(5))), mkt(6))},
	}

	for _, te := range tests {
		if s := Size(te.tree); s != te.size {
			t.Error("Expected size", te.size, "but computed size was", s)
		}
	}
}

func TestEnumerate(t *testing.T) {
	t5 := &nod{6, nil}
	t4 := &nod{5, nil}
	t3 := &nod{4, []*nod{t4, t5}}
	t2 := &nod{3, nil}
	t1 := &nod{2, []*nod{t3, t2}}
	t0 := &nod{1, nil}
	tree := &nod{0, []*nod{t0, t1}}
	if t5.ChiCo() != 0 {
		t.Error("Wrong children count")
	}
	t.Log(tree)
	r, d, h := Enumerate(tree)
	t.Log("Tree", r)
	t.Log("Depth", d)
	t.Log("Height", h)
	t.Fail()
}

func TestTraverse(t *testing.T) {
	t5 := &nod{6, nil}
	t4 := &nod{5, nil}
	t3 := &nod{4, []*nod{t4, t5}}
	t2 := &nod{3, nil}
	t1 := &nod{2, []*nod{t3, t2}}
	t0 := &nod{1, nil}
	tree := &nod{0, []*nod{t0, t1}}

	curDep := 0
	depths := list.New()
	items := list.New()
	heights := make(map[Tree]int)
	f1 := func(r Tree) {
		// Save current depth of the node and prepare for next child
		depths.PushBack(curDep)
		curDep++
		// Save item in the list
		items.PushBack(r)
	}
	f2 := func(r Tree) {
		// Children visited, go back
		curDep--
		// Node height is the maximum height of children + 1, if any
		mhc := 0
		if r.ChiCo() > 0 {
			for i := 0; i < r.ChiCo(); i++ {
				c := r.Child(i)
				if heights[c] > mhc {
					mhc = heights[c]
				}
			}
			mhc++
		}
		heights[r] = mhc
	}

	Traverse(tree, f1, f2)

	for e, f := depths.Front(), items.Front(); e != nil; e, f = e.Next(), f.Next() {
		println(e.Value.(int), " ", f.Value.(*nod).String(), " ", heights[f.Value.(*nod)])
	}
}

func TestTraverseParent(t *testing.T) {
	t6 := &nod{6, nil}
	t5 := &nod{5, nil}
	t4 := &nod{4, []*nod{t5, t6}}
	t3 := &nod{3, nil}
	t2 := &nod{2, []*nod{t4, t3}}
	t1 := &nod{1, nil}
	t0 := &nod{0, []*nod{t1, t2}}

	// Enumerate the tree and get node parents and depths
	curDep, depths := 0, make([]int, 0) // Depth of each node

	parents := make([]int, 0)        // Index of the parent of each node
	curInd, parInd := -1, list.New() // Stores the node indices of parents
	parInd.PushBack(curInd)          // Start with non existing node

	ent := func(n Tree) {
		depths = append(depths, curDep)
		curDep++

		curInd++                         // Increase node index/counter
		ind := parInd.Back().Value.(int) // Get index of current parent
		parents = append(parents, ind)   // Save parent index
		parInd.PushBack(curInd)          // Set current node as new parent

	}
	exi := func(n Tree) {
		curDep--

		parInd.Remove(parInd.Back()) // Remove node from list of parents
	}
	nodes := Traverse(t0, ent, exi)

	// Iterate every node and check if their parents are correct

	t.Log("Computed parents", parents)
	t.Log("Computed nodes:", nodes)
	t.Log("Computed depths", depths)
	t.Fail()
}
