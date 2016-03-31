package base

import (
	"github.com/akiross/gogp/image/draw2d/imgut"
	"github.com/gonum/floats"
	"math"
)

func fitnessRMSE(ind, targ *imgut.Image) float64 {
	// Images to vector
	dataInd := imgut.ToSlice(ind)
	dataTarg := imgut.ToSlice(targ)
	// (root mean square) error
	floats.Sub(dataInd, dataTarg)
	// (root mean) square error
	floats.Mul(dataInd, dataInd)
	// (root) mean square error
	totErr := floats.Sum(dataInd)
	return math.Sqrt(totErr / float64(len(dataInd)))
}

func fitnessRMSEImage(ind, targ *imgut.Image) float64 {
	return imgut.PixelRMSE(ind, targ)
}

func fitnessLinearScalingRMSE(ind, targ *imgut.Image) float64 {
	return 0.0
}
