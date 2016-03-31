package main

import (
	"flag"
	"fmt"
	"github.com/akiross/gogp/apps/base"
	"github.com/akiross/gogp/apps/base/repr/rr"
	"github.com/akiross/gogp/apps/evolve"
	"github.com/akiross/gogp/image/draw2d/imgut"
	rrepr "github.com/akiross/gogp/repr/rr"
	"math/rand"
	"os"
	"strings"
)

func draw(ind *base.Individual, img *imgut.Image) {
	rr.Draw(ind.Node, img)
}

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fPalSolid := fs.Bool("pf", false, "Enable Palette Full colors")
	fPalShade := fs.Bool("ps", false, "Enable Palette Shade colors")
	fEphFull := fs.Bool("ef", false, "Enable Full-color Ephemerals")
	fEphShade := fs.Bool("es", false, "Enable randomly-oriented Shaded-color Ephemerals")
	fEphDiagFill := fs.Bool("edf", false, "Enable Diagonal-oriented Full-color Ephemerals")
	fEphDiagLine := fs.Bool("edl", false, "Enable Diagonal-oriented Line-color Ephemerals")
	maxDepth := fs.Int("maxdepth", 13, "Set the maximum depth (default 13)")
	fs.Parse(os.Args[1:])

	// After parsing, change the name of the program to reflect used flags
	newName := strings.Join(os.Args[:len(os.Args)-fs.NArg()], " ")
	// Prepare arguments for next stage
	os.Args = append([]string{newName}, fs.Args()...)

	// Enable terminals according to flags
	if *fEphFull {
		rr.Terminals = append(rr.Terminals, rrepr.MakeEphimeral("MakeFull", rr.MakeFullColor))
	}
	if *fEphShade {
		rr.Terminals = append(rr.Terminals, rrepr.MakeEphimeral("MakeShade", rr.MakeShadeColor))
	}
	if *fEphDiagFill {
		rr.Terminals = append(rr.Terminals, rrepr.MakeEphimeral("MakeDiagFill", rr.MakeDiagFill))
	}
	if *fEphDiagLine {
		rr.Terminals = append(rr.Terminals, rrepr.MakeEphimeral("MakeDiagLine", rr.MakeDiagLine))
	}
	if *fPalSolid {
		count := 16 // Number of total colors, from black to white
		for i := 0; i < count; i++ {
			c := float64(i) / float64(count-1)
			name := fmt.Sprintf("G%02X", int(c*255))
			rr.Terminals = append(rr.Terminals, rrepr.MakeTerminal(name, rrepr.Filler(c, c, c, 1)))
		}
	}

	// Build some shades
	/* shading positions are limited to a grid of discrete end/start points.
	a     b     c       For example, with 3 points on each axis, we are
	|     |     |       limited to 9 possible starting and ending points
	+-----+-----+- a    in this way, we reduce the space of terminals
	|           |       and make easier to identify each terminal with a
	|           +- b    simple, textual name, for example
	|           |       LFaaEbc means that we have a shading from white (F)
	+-----------+- c    to light gray (E) from position (a, a) to position (b, c)
	*/
	if *fPalShade {
		count := 8
		reps := 8
		// TODO we didn't implement the strategy above yet.
		//sidePoints := 8
		for i := 0; i <= count; i++ {
			c := float64(i) / float64(count)
			for j := i + 1; j <= count; j++ {
				k := float64(j) / float64(count)
				// Multiple copie
				for n := 0; n < reps; n++ {
					sx, sy, ex, ey := rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64()
					name := fmt.Sprintf("L_%d-%d_%d-%d_%d-%d", int(c*255), int(k*255), int(sx*100), int(sy*100), int(ex*100), int(ey*100))
					rr.Terminals = append(rr.Terminals, rrepr.MakeTerminal(name, rrepr.LinShade(c, k, sx, sy, ex, ey)))
				}
			}
		}

	}

	// Fallback on black and white
	if len(rr.Terminals) == 0 {
		rr.Terminals = append(rr.Terminals, rrepr.MakeTerminal("Black", rrepr.Filler(0, 0, 0, 1)))
		rr.Terminals = append(rr.Terminals, rrepr.MakeTerminal("White", rrepr.Filler(1, 1, 1, 1)))
	}
	// Run second phase
	evolve.Evolve(rr.MakeMaxDepth(*maxDepth), rr.Functionals, rr.Terminals, draw)
}
