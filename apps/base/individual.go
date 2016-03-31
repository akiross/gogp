package base

import (
	"fmt"
	"github.com/akiross/gogp/ga"
	"github.com/akiross/gogp/gp"
	"github.com/akiross/gogp/image/draw2d/imgut"
	"github.com/akiross/gogp/node"
	"github.com/akiross/gogp/util/stats/sequence"
	"github.com/gonum/floats"
	"math"
)

type Settings struct {
	// We keep this low, because trees may grow too large
	// and use too much memory
	MaxDepth int

	// Images used for evaluation
	//ImgTarget, ImgTemp *imgut.Image
	ImgTarget *imgut.Image

	// Functionals and terminals used
	Functionals []gp.Primitive
	Terminals   []gp.Primitive

	Draw func(*Individual, *imgut.Image)

	// Operators used in evolution
	CrossOver func(float64, *Individual, *Individual) bool
	Mutate    func(float64, *Individual) bool

	// These hold general purpose statistics for debugging purposes
	Statistics map[string]*sequence.SequenceStats // Custom stats
}

type Individual struct {
	Node       *node.Node
	fitness    ga.Fitness
	fitIsValid bool
	set        *Settings
	ImgTemp    *imgut.Image // where to render the individual
}

func (ind *Individual) String() string {
	return fmt.Sprint(ind.Node)
}

func (ind *Individual) MarshalJSON() ([]byte, error) {
	return ind.Node.MarshalJSON()
}

func (ind *Individual) Fitness() ga.Fitness {
	if !ind.fitIsValid {
		return ind.Evaluate()
	} else {
		return ind.fitness
	}
}

func (ind *Individual) Copy() ga.Individual {
	tmpImg := imgut.Create(ind.set.ImgTarget.W, ind.set.ImgTarget.H, ind.set.ImgTarget.ColorSpace)
	return &Individual{ind.Node.Copy(), ind.fitness, ind.fitIsValid, ind.set, tmpImg}
}

func (ind *Individual) Crossover(pCross float64, mate ga.Individual) {
	if ind.set.CrossOver(pCross, ind, mate.(*Individual)) {
		ind.Invalidate()
		mate.(*Individual).Invalidate()
	}
}

func (ind *Individual) Draw(img *imgut.Image) {
	ind.set.Draw(ind, img)
}

// This method sould always evaluate, eventually saving to cache
func (ind *Individual) Evaluate() ga.Fitness {
	// Draw the individual
	ind.set.Draw(ind, ind.ImgTemp)

	var rmse float64

	if false { // Linear scaling
		// Images to vector
		dataInd := imgut.ToSlice(ind.ImgTemp)
		dataTarg := imgut.ToSlice(ind.set.ImgTarget)
		// Compute average pixels
		avgy := floats.Sum(dataInd) / float64(len(dataInd))
		avgt := floats.Sum(dataTarg) / float64(len(dataTarg))

		// Difference y - avgy
		y_avgy := make([]float64, len(dataInd))
		copy(y_avgy, dataInd)
		floats.AddConst(-avgy, y_avgy)
		// Difference t - avgt
		t_avgt := make([]float64, len(dataTarg))
		copy(t_avgt, dataTarg)
		floats.AddConst(-avgt, t_avgt)
		// Multuplication (t - avgt)(y - avgy)
		floats.Mul(t_avgt, y_avgy)
		// Summation
		numerator := floats.Sum(t_avgt)
		// Square (y - avgy)^2
		floats.Mul(y_avgy, y_avgy)
		denomin := floats.Sum(y_avgy)
		// Compute b-value
		b := numerator / denomin
		// Compute a-value
		a := avgt - b*avgy

		// Compute now the scaled RMSE, using y' = a + b*y
		floats.Scale(b, dataInd)      // b*y
		floats.AddConst(a, dataInd)   // a + b*y
		floats.Sub(dataInd, dataTarg) // (a + b * y - t)
		floats.Mul(dataInd, dataInd)  // (a + b * y - t)^2
		total := floats.Sum(dataInd)  // Sum(...)
		rmse = math.Sqrt(total / float64(len(dataInd)))

		// Save RMSE as fitness
		ind.fitness = ga.Fitness(rmse)
	} else { // Normal RMSE image
		// Compute RMSE
		rmse = imgut.PixelRMSE(ind.ImgTemp, ind.set.ImgTarget)
		ind.fitness = ga.Fitness(rmse)
	}

	// When true, it will multiply by edge-detection RMSE
	if false {
		// Compute edge detection
		edgeKern := &imgut.ConvolutionMatrix{3, []float64{
			0, 1, 0,
			1, -4, 1,
			0, 1, 0},
		}
		imgEdge := imgut.ApplyConvolution(edgeKern, ind.ImgTemp)
		targEdge := imgut.ApplyConvolution(edgeKern, ind.set.ImgTarget) // XXX this could be computed once
		// Compute distance between edges
		edRmse := imgut.PixelRMSE(imgEdge, targEdge)

		// Statistics on output values
		if _, ok := ind.set.Statistics["sub-fit-plain"]; !ok {
			ind.set.Statistics["sub-fit-plain"] = sequence.Create()
		}
		ind.set.Statistics["sub-fit-plain"].Observe(rmse)

		if _, ok := ind.set.Statistics["sub-fit-edged"]; !ok {
			ind.set.Statistics["sub-fit-edged"] = sequence.Create()
		}
		ind.set.Statistics["sub-fit-edged"].Observe(edRmse)
		// Weighted fitness
		ind.fitness = ga.Fitness(rmse * edRmse)
	}
	ind.fitIsValid = true
	return ind.fitness
}

func (ind *Individual) FitnessValid() bool {
	return ind.fitIsValid
}

func (ind *Individual) Invalidate() {
	ind.fitIsValid = false
}

func (ind *Individual) Initialize() {
	ind.Node = node.MakeTreeHalfAndHalf(ind.set.MaxDepth, ind.set.Functionals, ind.set.Terminals)
	ind.ImgTemp = imgut.Create(ind.set.ImgTarget.W, ind.set.ImgTarget.H, ind.set.ImgTarget.ColorSpace)
}

func (ind *Individual) Mutate(pMut float64) {
	if ind.set.Mutate(pMut, ind) {
		ind.Invalidate()
	}
}
