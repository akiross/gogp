/*
Package gogp provides the foundation to use Genetic Programming (GP) in Go.

The package provides basic functions and methods to implement a GP system, such as:

- Node mutation

- One point crossover

The code is organized around trees, therefore only tree-like structures are admitted.

To use gogp, you have to create the primitives (functionals and terminals) that will
compose your solutions.

1. Create a type for terminals: this type usually identifies the number of variables
that you are using as input (i.e. how many input variables do you have?)

2. Create a type for every functional group that you are about to use: each group is
characterized by the same number of parameters (arity). For example, if you use
unary and binary functionals, you need two types.

3. Implement your operators.

*/
package gp

import (
	"reflect"
	"runtime"
	"strings"
)

type Primitive interface {
	IsFunctional() bool         // Returns true if is functional, false if terminal
	Arity() int                 // Returns the arity of the primitive
	Run(...Primitive) Primitive // Functionals returns terminals, terminals do nothing
	Name() string               // Get the name of this primitive
}

func FuncName(f Primitive) string {
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	s := strings.Split(name, ".")
	return s[len(s)-1]
}
