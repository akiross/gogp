package main

import (
	"github.com/akiross/gogp/apps/base"
	"github.com/akiross/gogp/apps/base/repr/rr"
	"github.com/akiross/gogp/apps/evolve"
	"github.com/akiross/gogp/image/draw2d/imgut"
)

func draw(ind *base.Individual, img *imgut.Image) {
	rr.Draw(ind.Node, img)
}

func main() {
	evolve.Evolve(rr.MaxDepth, rr.Functionals, rr.Terminals, draw)
}
