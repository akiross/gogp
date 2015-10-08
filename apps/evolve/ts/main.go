package main

import (
	"github.com/akiross/gogp/apps/base"
	"github.com/akiross/gogp/apps/base/repr/ts"
	"github.com/akiross/gogp/apps/evolve"
	"github.com/akiross/gogp/image/draw2d/imgut"
)

func draw(ind *base.Individual, img *imgut.Image) {
	ts.Draw(ind.Node, img)
}

func main() {
	evolve.Evolve(ts.MaxDepth, ts.Functionals, ts.Terminals, draw)
}
