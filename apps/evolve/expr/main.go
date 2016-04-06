package main

import (
	"flag"
	"github.com/akiross/gogp/apps/base"
	"github.com/akiross/gogp/apps/base/repr/expr"
	"github.com/akiross/gogp/apps/evolve"
	"github.com/akiross/gogp/image/draw2d/imgut"
	"github.com/akiross/gogp/repr/expr/binary"
	"math"
	"math/rand"
	"os"
	"strings"
)

func draw(ind *base.Individual, img *imgut.Image) {
	expr.Draw(ind.Node, img)
}

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	//fTrig := fs.Bool("tf", false, "Enable sin, cos, tanh")
	maxDepth := fs.Int("maxdepth", 13, "Set the maximum depth (default 13)")
	fs.Parse(os.Args[1:])

	// After parsing, change the name of the program to reflect used flags
	newName := strings.Join(os.Args[:len(os.Args)-fs.NArg()], " ")
	// Prepare arguments for next stage
	os.Args = append([]string{newName}, fs.Args()...)

	expr.Functionals = append(expr.Functionals, binary.MakeTernary("ITE", func(a, b, c binary.NumericOut) binary.NumericOut {
		if a >= 0 {
			return b
		} else {
			return c
		}
	}))
	expr.Functionals = append(expr.Functionals, binary.MakeBinary("Sum", func(a, b binary.NumericOut) binary.NumericOut { return a + b }))
	expr.Functionals = append(expr.Functionals, binary.MakeBinary("Sub", func(a, b binary.NumericOut) binary.NumericOut { return a - b }))
	expr.Functionals = append(expr.Functionals, binary.MakeBinary("Mul", func(a, b binary.NumericOut) binary.NumericOut { return a * b }))
	expr.Functionals = append(expr.Functionals, binary.MakeBinary("Div", func(a, b binary.NumericOut) binary.NumericOut {
		if b == 0 {
			return binary.NumericOut(1)
		} else {
			return a / b
		}
	}))
	expr.Functionals = append(expr.Functionals, binary.MakeBinary("Min", func(a, b binary.NumericOut) binary.NumericOut {
		if a < b {
			return a
		} else {
			return b
		}
	}))
	expr.Functionals = append(expr.Functionals, binary.MakeBinary("Max", func(a, b binary.NumericOut) binary.NumericOut {
		if a > b {
			return a
		} else {
			return b
		}
	}))
	expr.Functionals = append(expr.Functionals, binary.MakeBinary("Pow", func(a, b binary.NumericOut) binary.NumericOut {
		return binary.NumericOut(math.Pow(float64(a), float64(b)))
	}))

	expr.Functionals = append(expr.Functionals, binary.MakeUnary("Sqr", func(a binary.NumericOut) binary.NumericOut { return a * a }))
	expr.Functionals = append(expr.Functionals, binary.MakeUnary("Sqrt", func(a binary.NumericOut) binary.NumericOut {
		if a < 0 {
			return binary.NumericOut(math.Sqrt(float64(-a)))
		} else {
			return binary.NumericOut(math.Sqrt(float64(a)))
		}
	}))
	expr.Functionals = append(expr.Functionals, binary.MakeUnary("Abs", func(a binary.NumericOut) binary.NumericOut {
		if a < 0 {
			return -a
		} else {
			return a
		}
	}))
	expr.Functionals = append(expr.Functionals, binary.MakeUnary("Neg", func(a binary.NumericOut) binary.NumericOut { return -a }))
	expr.Functionals = append(expr.Functionals, binary.MakeUnary("Sign", func(a binary.NumericOut) binary.NumericOut {
		if a < 0 {
			return -1
		} else {
			return 1
		}
	}))
	expr.Functionals = append(expr.Functionals, binary.MakeUnary("Sin", func(a binary.NumericOut) binary.NumericOut {
		return binary.NumericOut(math.Sin(float64(a)))
	}))
	expr.Functionals = append(expr.Functionals, binary.MakeUnary("Cos", func(a binary.NumericOut) binary.NumericOut {
		return binary.NumericOut(math.Cos(float64(a)))
	}))
	expr.Functionals = append(expr.Functionals, binary.MakeUnary("Tanh", func(a binary.NumericOut) binary.NumericOut {
		return binary.NumericOut(math.Tanh(float64(a)))
	}))

	expr.Terminals = append(expr.Terminals, binary.MakeIdentityX())
	expr.Terminals = append(expr.Terminals, binary.MakeIdentityY())
	expr.Terminals = append(expr.Terminals, binary.MakeConstant(-1))
	expr.Terminals = append(expr.Terminals, binary.MakeConstant(0))
	expr.Terminals = append(expr.Terminals, binary.MakeConstant(1))
	expr.Terminals = append(expr.Terminals, binary.MakeConstant(2))
	expr.Terminals = append(expr.Terminals, binary.MakeConstant(10))

	expr.Terminals = append(expr.Terminals, binary.MakeEphimeral("MakeRand", func() *binary.Primitive {
		v := rand.Float64()
		return binary.MakeConstant(binary.NumericOut(v))
	}))

	// Run second phase
	evolve.Evolve(expr.MakeMaxDepth(*maxDepth), expr.Functionals, expr.Terminals, draw)
}
