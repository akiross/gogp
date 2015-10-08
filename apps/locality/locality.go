package main

import (
	"fmt"
	"github.com/akiross/gogp/apps/base/repr/expr"
	"github.com/akiross/gogp/apps/base/repr/ts"
	"github.com/akiross/gogp/apps/base/repr/vhs"
	"github.com/akiross/gogp/image/draw2d/imgut"
	"github.com/akiross/gogp/node"
	"math/rand"
	"time"
)

const (
	N     = 20
	M     = 20
	MAX_D = 8
	IMG_W = 100
	IMG_H = 100
)

// Locality test
func testRepr(initFunc func(maxDep int) *node.Node, mutateFunc func(float64, *node.Node), paintFunc func(*node.Node, *imgut.Image)) float64 {
	// Create storage for the images
	indImage := imgut.Create(IMG_W, IMG_H, imgut.MODE_RGB)
	tmpImage := imgut.Create(IMG_W, IMG_H, imgut.MODE_RGB)

	// Build random individuals
	randomIndividuals := make([]*node.Node, N)
	// For each individual
	var exprError float64 = 0
	for _, i := range randomIndividuals {
		// Initialize it
		i = initFunc(MAX_D)
		// Render it to an image (it will be garbage)
		indImage.Clear()
		paintFunc(i, indImage)
		// Average error for this individual
		var indErrorAvg float64 = 0
		var indErrorVar float64 = 0
		for k := 0; k < M; k++ {
			// Copy the individual
			j := i.Copy()
			// Mutate the individual
			mutateFunc(1, j)

			//////////////////////////////////////////////////
			// FIXME TODO this should be removed in production: test that original is not changed
			tmpImage.Clear()
			paintFunc(i, tmpImage)
			// Check distance
			d := imgut.PixelDistance(indImage, tmpImage)
			if d != 0 {
				fmt.Println("ERROR! Distance of i to itself is not zero!", d)
				panic("STOPPING TO DEBUG")
			}
			//////////////////////////////////////////////////

			// Render it to another image
			tmpImage.Clear()
			paintFunc(j, tmpImage)
			// Compute distance
			dist := imgut.PixelRMSE(indImage, tmpImage)
			// Accumulate distance
			indErrorAvg += dist
			indErrorVar += dist * dist
		}
		// Compute average error
		indErrorAvg = indErrorAvg / float64(M)
		// Compute variance
		indErrorVar = indErrorVar - (indErrorAvg * indErrorAvg)
		fmt.Println("Individual avg error and variance:", indErrorAvg, indErrorVar)

		// Accumulate
		exprError += indErrorAvg
	}
	return exprError / float64(N)
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	// For each representation
	//   Build N individuals (e.g. 10000)
	//   For each individual i
	//     Mutate i M times (e.g. 50), yielding j_k
	//     Compute avg (over k) distance between i and j_k

	// The initialization function creates a tree with specified max depth
	// And it's a closure over the primitives
	exprInit := func(maxDep int) *node.Node {
		return node.MakeTreeHalfAndHalf(maxDep, expr.Functionals, expr.Terminals)
	}
	exprMutate := node.MakeSubtreeMutation(MAX_D, exprInit)
	fmt.Println("Testing representation: expr")
	exprError := testRepr(exprInit, exprMutate, expr.Draw)
	fmt.Println("Expr total error:", exprError)

	tsInit := func(maxDep int) *node.Node {
		return node.MakeTreeHalfAndHalf(maxDep, ts.Functionals, ts.Terminals)
	}
	tsMutate := node.MakeSubtreeMutation(MAX_D, tsInit)
	fmt.Println("Testing representation: TS")
	tsError := testRepr(tsInit, tsMutate, ts.Draw)
	fmt.Println("TS total error:", tsError)

	vhsInit := func(maxDep int) *node.Node {
		return node.MakeTreeHalfAndHalf(maxDep, vhs.Functionals, vhs.Terminals)
	}
	vhsMutate := node.MakeSubtreeMutation(MAX_D, vhsInit)
	fmt.Println("Testing representation: VHS")
	vhsError := testRepr(vhsInit, vhsMutate, vhs.Draw)
	fmt.Println("VHS total error:", vhsError)

}
