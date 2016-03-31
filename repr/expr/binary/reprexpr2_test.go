package binary

import (
	"github.com/akiross/gogp/image/draw2d/imgut"
	"math"
	"math/rand"
	"testing"
)

func TestPrimitives(t *testing.T) {
	x := MakeIdentityX()
	y := MakeIdentityY()
	c5 := MakeConstant(0.5)
	e := MakeEphimeral("MakeRand", func() *Primitive {
		v := rand.Float64()
		return MakeConstant(NumericOut(v))
	})

	sum := MakeBinary("Sum", func(a, b NumericOut) NumericOut { return a + b })
	sub := MakeBinary("Sub", func(a, b NumericOut) NumericOut { return a - b })
	mul := MakeBinary("Mul", func(a, b NumericOut) NumericOut { return a * b })

	cos := MakeUnary("Cos", func(a NumericOut) NumericOut { return NumericOut(math.Cos(float64(a))) })

	spl := MakeTernary("Spl", func(a, b, c NumericOut) NumericOut {
		if a > 0.5 {
			return b
		} else {
			return c
		}
	})

	sub = sub
	sum = sum
	mul = mul
	cos = cos
	spl = spl
	e, x, y, c5 = e, x, y, c5

	expr := spl.Run(x, y, x)
	t.Log(expr)

	img := imgut.Create(100, 100, imgut.MODE_RGBA)

	// We have to compile the nodes
	exec := expr.Run().(*Primitive)
	// Apply the function
	//	exec(0 0, float64(img.W), float64(img.H), img)
	var call imgut.PixelFunc = func(x, y float64) float64 {
		return float64(exec.Eval(NumericIn(x), NumericIn(y)))
	}
	img.FillMathBounds(call)

	img.WritePNG("test_expr_repr.png")
	t.Fail()
}
