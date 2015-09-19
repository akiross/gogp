package expr

import "ale-re.net/phd/gogp"

func Identity1(x int) int {
	return x
}

func Constant1(c int) Terminal1 {
	return func(_ int) int {
		return c
	}
}

func Constant2(c int) Terminal2 {
	return func(_, _ int) int {
		return c
	}
}

func Sum(args ...gogp.Primitive) gogp.Primitive {
	return Terminal1(func(x int) int {
		return args[0].(Terminal1)(x) + args[1].(Terminal1)(x)
	})
}

func Sub(args ...gogp.Primitive) gogp.Primitive {
	return Terminal1(func(x int) int {
		return args[0].(Terminal1)(x) - args[1].(Terminal1)(x)
	})
}

func Abs(args ...gogp.Primitive) gogp.Primitive {
	return Terminal1(func(x int) int {
		v := args[0].(Terminal1)(x)
		if v < 0 {
			return -v
		} else {
			return v
		}
	})
}
