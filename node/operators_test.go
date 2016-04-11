package node

import (
	"fmt"
	"github.com/akiross/gogp/gp"
	"math/rand"
	"testing"
	"time"
)

func c_zero(_ int) int { return 0 }
func c_one(_ int) int  { return 1 }

func genBalTree(maxDepth int) *Node {
	return MakeTreeGrowBalanced(maxDepth, []gp.Primitive{
		Functional2(Sum),
		Functional2(Sub),
		Functional1(Abs),
	}, []gp.Primitive{
		Terminal1(c_zero),
		Terminal1(c_one),
		Terminal1(Identity1),
	})
}

func init() {
	rand.Seed(time.Now().Unix())
}

func TestMakeTreeGrowBalanced(t *testing.T) {
	// TODO how do I check if the trees are really what I am expecting?
	numTerms, numFuncs := 0, 0
	for i := 0; i < 10000; i++ {
		t := genBalTree(10)
		nods, _, _ := t.Enumerate()
		for _, v := range nods {
			if v.value.Arity() > 0 {
				numFuncs++
			} else {
				numTerms++
			}
		}
	}
	t.Error("Termals vs functionals", float64(numTerms)/float64(numFuncs))
}

func TestMakeTreeFull(t *testing.T) {
	// Generates trees of random height with full
	// enumerate the nodes and count how many are
	// at a given depth. They should be perfectly
	// distributed for a binary tree.

	for d := 0; d < 10; d++ {
		counts := make(map[int]int)

		tr := MakeTreeFull(d, []gp.Primitive{
			Functional2(Sum),
			Functional2(Sub),
		}, []gp.Primitive{
			Terminal1(c_zero),
			Terminal1(c_one),
		})
		// Count how many nodes for each depth
		_, tdep, _ := tr.Enumerate()
		for _, j := range tdep {
			counts[j] += 1
		}
		// Check if the number of nodes is correct
		for j, n := 0, 1; j < d; j, n = j+1, n*2 {
			if counts[j] != n {
				t.Error("at level", j, "we expected", n, "nodes, but", counts[j], "were found")
			}
		}
	}
}

func TestSubtreeMutation(t *testing.T) {

}

func TestPartitionLeaves(t *testing.T) {
	zero, one, id := Terminal1(Constant1(0)), Terminal1(Constant1(1)), Terminal1(Identity1)
	sum, abs, sub := Functional2(Sum), Functional1(Abs), Functional2(Sub)
	var tZero, tOne, tId *Node = mt(zero), mt(one), mt(id) //mt(sum, mt(one), mt(abs, mt(sub, mt(zero), mt(id))))

	t1 := mt(sum, mt(abs, tId), mt(sub, tZero, tOne))

	nods, _, _ := t1.Enumerate()

	le, nonle := partitionLeaves(nods)

	t.Log("Leaves:", le)
	t.Log("Non-leaves:", nonle)

	for i := range le {
		j := le[i]
		if len(nods[j].children) != 0 {
			t.Error("Leaf expected, but had child(ren)", j, nods[j], nods[j].children)
		}
	}
	for i := range nonle {
		j := nonle[i]
		if len(nods[j].children) == 0 {
			t.Error("Non-leaf expected, but didn't have child(ren)", i, j, nods[j])
		}
	}
}

func TestSubtreeMutationLevelExp(t *testing.T) {
	zero, one, id := Terminal1(c_zero), Terminal1(c_one), Terminal1(Identity1)
	sum, abs, sub := Functional2(Sum), Functional1(Abs), Functional2(Sub)
	var tZero, tOne, tId *Node = mt(zero), mt(one), mt(id) //mt(sum, mt(one), mt(abs, mt(sub, mt(zero), mt(id))))

	t1 := mt(sum, mt(sub, tZero, mt(sum, tOne, tOne)), mt(abs, tId))

	nods, depths, _ := t1.Enumerate()

	le, nonle := partitionLeaves(nods)

	t.Log("Leaves:", le)
	t.Log("Non-leaves:", nonle)
	probs, index := makeExpProbs(depths, le, 2)
	_, _ = probs, index
	//	t.Error("Tree\n", t1.PrettyPrint())
	//	t.Error("Probs", probs)
	//	t.Error("Index", index)
}

func TestSubtreeMutationGuided(t *testing.T) {
	// TODO testare questa
}

func TestArityDepthProbComputer(t *testing.T) {
	// Generate a random tree
	tr := genBalTree(3)
	nodes, _, _ := tr.Enumerate()

	// Build probability computer
	nlf := ArityDepthProbComputer(tr)

	// Print the tree with probabs
	t.Log("\n" + PrettyPrint(tr, func(n Tree) string {
		nn := n.(*Node)
		return fmt.Sprintf("%v p=%v", nn.value.Name(), nlf(nn))
	}))

	// Build CDF
	cdf := make([]float64, len(nodes))
	inds := make([]int, len(nodes))
	for i, v := range nodes {
		inds[i] = i
		cdf[i] = nlf(v)
	}
	normalSlice(cdf)
	probs := computeCDFinPlace(cdf, inds) // Compute CDF slice

	// For a large number of time, extract a node according to distribution
	// and count how many times each one is extracted
	counts := make([]int, len(nodes))
	tot := 100000
	for i := 0; i < tot; i++ {
		nid := extractCFDinPlace(cdf) // Extract node index
		counts[nid]++
	}

	relFreq := make([]float64, len(nodes))
	for i, c := range counts {
		relFreq[i] = float64(c) / float64(tot)
	}

	// The counting should approximate the distribution
	t.Log(relFreq, probs, inds)

	t.Fail()
}
