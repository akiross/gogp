package node

import "testing"

func c_zero(_ int) int { return 0 }
func c_one(_ int) int  { return 1 }

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
}

func TestArityDepthProbComputer(t *testing.T) {
	zero, one, id := Terminal1(c_zero), Terminal1(c_one), Terminal1(Identity1)
	sum, abs, sub := Functional2(Sum), Functional1(Abs), Functional2(Sub)
	var tZero, tOne, tId *Node = mt(zero), mt(one), mt(id) //mt(sum, mt(one), mt(abs, mt(sub, mt(zero), mt(id))))

	t1 := mt(sum, mt(sub, tZero, mt(sum, tOne, tOne)), mt(abs, tId))

	t.Log("\n", t1.PrettyPrint())

	nlf := ArityDepthProbComputer(t1)

	for i, n := range Traverse(t1, nil, nil) {
		t.Log("NLF", i, n, nlf(n.(*Node)))
	}
	t.Fail()
}
