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
	N     = 1000
	M     = 50
	MAX_D = 8
	IMG_W = 100
	IMG_H = 100
)

// Locality test
func testRepr(initFunc func(maxDep int) *node.Node, mutateFunc func(float64, *node.Node), paintFunc func(*node.Node, *imgut.Image)) (avgErr, varErr float64) {
	// Create storage for the images
	indImage := imgut.Create(IMG_W, IMG_H, imgut.MODE_RGB)
	tmpImage := imgut.Create(IMG_W, IMG_H, imgut.MODE_RGB)

	// Build random individuals
	randomIndividuals := make([]*node.Node, N)
	// For each individual
	var totErrorSum float64 = 0
	var totErrorSqr float64 = 0
	for _, i := range randomIndividuals {
		// Initialize it
		i = initFunc(MAX_D)
		// Render it to an image (it will be garbage)
		indImage.Clear()
		paintFunc(i, indImage)
		// Average error for this individual
		var indErrorSum float64 = 0
		var indErrorSqr float64 = 0
		for k := 0; k < M; k++ {
			// Copy the individual
			j := i.Copy()
			// Mutate the individual
			mutateFunc(1, j)
			// Render it to another image
			tmpImage.Clear()
			paintFunc(j, tmpImage)
			// Compute distance
			dist := imgut.PixelRMSE(indImage, tmpImage)
			// Accumulate distance
			indErrorSum += dist
			indErrorSqr += dist * dist
		}
		// Compute average error
		indErrorAvg := indErrorSum / float64(M)
		// Compute variance
		indErrorVar := indErrorSqr/float64(M) - (indErrorAvg * indErrorAvg)
		fmt.Println("  Individual avg error and variance:", indErrorAvg, indErrorVar)

		// Accumulate error
		totErrorSum += indErrorAvg
		totErrorSqr += indErrorAvg * indErrorAvg
	}
	// Compute average error of the averages
	avgErr = totErrorSum / float64(N)
	// Compute variance of the averages
	varErr = totErrorSqr/float64(N) - (avgErr * avgErr)

	fmt.Println("Total error and var:", avgErr, varErr)
	return
}

func main() {
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	fmt.Println("Using seed:", seed)

	// Test expr
	// The initialization function creates a tree with specified max depth
	// And it's a closure over the primitives
	exprInit := func(maxDep int) *node.Node {
		return node.MakeTreeHalfAndHalf(maxDep, expr.Functionals, expr.Terminals)
	}
	exprMutate := node.MakeSubtreeMutation(MAX_D, exprInit)
	fmt.Println("Testing representation: expr")
	exprErrorAvg, exprErrorVar := testRepr(exprInit, exprMutate, expr.Draw)
	fmt.Println("expr error avg:", exprErrorAvg, "var:", exprErrorVar)

	// Test ts
	tsInit := func(maxDep int) *node.Node {
		return node.MakeTreeHalfAndHalf(maxDep, ts.Functionals, ts.Terminals)
	}
	tsMutate := node.MakeSubtreeMutation(MAX_D, tsInit)
	fmt.Println("Testing representation: TS")
	tsErrorAvg, tsErrorVar := testRepr(tsInit, tsMutate, ts.Draw)
	fmt.Println("TS error avg:", tsErrorAvg, "var:", tsErrorVar)

	// Test vhs
	vhsInit := func(maxDep int) *node.Node {
		return node.MakeTreeHalfAndHalf(maxDep, vhs.Functionals, vhs.Terminals)
	}
	vhsMutate := node.MakeSubtreeMutation(MAX_D, vhsInit)
	fmt.Println("Testing representation: VHS")
	vhsErrorAvg, vhsErrorVar := testRepr(vhsInit, vhsMutate, vhs.Draw)
	fmt.Println("VHS error avg:", vhsErrorAvg, "var:", vhsErrorVar)
}
