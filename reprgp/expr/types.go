package expr

import "ale-re.net/phd/gogp"

type Terminal1 func(x int) int    // For expressions in 1 variable
type Terminal2 func(x, y int) int // For expressions in 2 variables

type Functional1 func(args ...gogp.Primitive) gogp.Primitive // Unary operations
type Functional2 func(args ...gogp.Primitive) gogp.Primitive // Binary operations

// The following are to satisfy the interface
func (self Terminal1) IsFunctional() bool                     { return false }
func (self Terminal1) Arity() int                             { return -1 }
func (self Terminal1) Run(p ...gogp.Primitive) gogp.Primitive { return self }

func (self Terminal2) IsFunctional() bool                     { return false }
func (self Terminal2) Arity() int                             { return -1 }
func (self Terminal2) Run(p ...gogp.Primitive) gogp.Primitive { return self }

func (self Functional1) IsFunctional() bool                     { return true }
func (self Functional1) Arity() int                             { return 1 }
func (self Functional1) Run(p ...gogp.Primitive) gogp.Primitive { return self(p[0]) }

func (self Functional2) IsFunctional() bool                     { return true }
func (self Functional2) Arity() int                             { return 2 }
func (self Functional2) Run(p ...gogp.Primitive) gogp.Primitive { return self(p[0], p[1]) }
