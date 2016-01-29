package main

import (
	"flag"
	"fmt"
	"github.com/akiross/gogp/apps/base/repr/rr"
	"github.com/akiross/gogp/gp"
	"github.com/akiross/gogp/image/draw2d/imgut"
	"github.com/akiross/gogp/node"
	"github.com/akiross/gogp/sa"
	"math/rand"
	"time"
)

type Solution struct {
	Node    *node.Node
	ImgTemp *imgut.Image // where to render the individual
	Conf    *Configuration
}

type Configuration struct {
	ImgTarget   *imgut.Image
	MaxDepth    int
	Functionals []gp.Primitive
	Terminals   []gp.Primitive
}

func (s *Solution) String() string {
	return fmt.Sprint(s.Node)
}

func (s *Solution) Copy() sa.Solution {
	tmpImg := imgut.Create(s.Conf.ImgTarget.W, s.Conf.ImgTarget.H, s.Conf.ImgTarget.ColorSpace)
	return &Solution{s.Node.Copy(), tmpImg, s.Conf}
}

func (s *Solution) BetterThan(sol sa.Solution) bool {
	myFit, solFit := s.Fitness(), sol.Fitness()
	return myFit < solFit
}

func (s *Solution) Mutate() {
	subtrMut := node.MakeSubtreeMutation(s.Conf.MaxDepth, func(maxDep int) *node.Node {
		return node.MakeTreeHalfAndHalf(maxDep, s.Conf.Functionals, s.Conf.Terminals)
	})
	subtrMut(s.Node)
}

func (s *Solution) Fitness() float64 {
	// Draw the individual
	rr.Draw(s.Node, s.ImgTemp)
	// Compute RMSE
	return imgut.PixelRMSE(s.ImgTemp, s.Conf.ImgTarget)
}

func (c *Configuration) RandomSolution() sa.Solution {
	n := node.MakeTreeHalfAndHalf(c.MaxDepth, c.Functionals, c.Terminals)
	tmpImg := imgut.Create(c.ImgTarget.W, c.ImgTarget.H, c.ImgTarget.ColorSpace)
	return &Solution{n, tmpImg, c}
}

func (c *Configuration) NeighborhoodSize() int {
	return 100
}

func (c *Configuration) MaxMoves() int {
	return 1000
}

func main() {
	targetPath := flag.String("t", "", "Target image (PNG) path")
	seed := flag.Int64("seed", time.Now().UTC().UnixNano(), "Seed for RNG")
	flag.Parse()

	rand.Seed(*seed)

	var conf Configuration
	// Solution max depth
	conf.MaxDepth = 13
	// Load the target image
	var err error
	conf.ImgTarget, err = imgut.Load(*targetPath)
	if err != nil {
		fmt.Println("ERROR: Cannot load image", *targetPath)
		panic("Cannot load image")
	}
	// Set terminals and functionals
	conf.Functionals = rr.Functionals
	conf.Terminals = rr.Terminals
	// Search with hill climbing
	sol := sa.CoinAnnealing(&conf)
	fmt.Println("Coin annealing done", sol, sol.Fitness())
	// Save solution somewhere
	outPath := "ca_best.png"
	fmt.Println("Saving best as", outPath)
	sol.Fitness()
	sol.(*Solution).ImgTemp.WritePNG(outPath)
}
