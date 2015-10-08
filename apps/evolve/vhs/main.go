package main

import (
	"github.com/akiross/gogp/apps/base"
	"github.com/akiross/gogp/apps/base/repr/vhs"
	"github.com/akiross/gogp/apps/evolve"
	"github.com/akiross/gogp/image/draw2d/imgut"
)

func draw(ind *base.Individual, img *imgut.Image) {
	vhs.Draw(ind.Node, img)
}

func main() {
	evolve.Evolve(vhs.MaxDepth, vhs.Functionals, vhs.Terminals, draw)
}
