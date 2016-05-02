package base

import (
	"bytes"
	"fmt"
	"github.com/akiross/gogp/ga"
	"github.com/akiross/gogp/gp"
	"github.com/akiross/gogp/image/draw2d/imgut"
	"github.com/akiross/gogp/node"
	"github.com/akiross/gogp/util/stats/counter"
	"github.com/akiross/gogp/util/stats/sequence"
	"github.com/gonum/floats"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Settings struct {
	// We keep this low, because trees may grow too large
	// and use too much memory
	MaxDepth int
	// Ramped initialization
	Ramped bool

	// Images used for evaluation
	//ImgTarget, ImgTemp *imgut.Image
	ImgTarget *imgut.Image

	// Functionals and terminals used
	Functionals []gp.Primitive
	Terminals   []gp.Primitive

	Draw func(*Individual, *imgut.Image)

	// Fitness func for one individual
	FitFunc func(ind *imgut.Image) float64

	// Operators used in evolution
	GenFunc   func(int) *node.Node // Generate tree
	Select    func([]*Individual, int) []ga.Individual
	CrossOver func(float64, *Individual, *Individual) bool
	Mutate    func(float64, *Individual) bool

	// These hold general purpose statistics for debugging purposes
	Statistics  map[string]*sequence.SequenceStats // Float values
	Counters    map[string]*counter.BoolCounter    // Count events
	IntCounters map[string]*counter.IntCounter     // Count ints

	// Minimization problem
	ga.MinProblem
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

// This method should always return the current futness
// Possibly caching evaluated results
func (ind *Individual) Fitness() ga.Fitness {
	if !ind.fitIsValid {
		ind.fitness = ind.Evaluate()
		ind.fitIsValid = true
	}
	return ind.fitness
}

// Returns true if ind is better than i
func (ind *Individual) BetterThan(i *Individual) bool {
	return ind.set.BetterThan(ind.Fitness(), i.Fitness())
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

func MakeFitMSE(targetImage *imgut.Image) func(*imgut.Image) float64 {
	dataTarg := imgut.ToSliceChans(targetImage, "R")
	return func(indImage *imgut.Image) float64 {
		// Get data
		dataImg := imgut.ToSliceChans(indImage, "R")
		// Difference (X - Y)
		floats.Sub(dataImg, dataTarg)
		// Squared (X - Y)^2
		floats.Mul(dataImg, dataImg)
		// Summation
		return floats.Sum(dataImg) / float64(len(dataImg))
	}
}

func MakeFitRMSE(targetImage *imgut.Image) func(*imgut.Image) float64 {
	// The MSE for compared
	msef := MakeFitMSE(targetImage)
	return func(indImage *imgut.Image) float64 {
		return math.Sqrt(msef(indImage))
	}
}

func MakeFitLinScale(targetImage *imgut.Image) func(*imgut.Image) float64 {
	// Pre-compute image to slice of floats
	dataTarg := imgut.ToSlice(targetImage)
	// Pre-compute average
	avgt := floats.Sum(dataTarg) / float64(len(dataTarg))
	return func(indImage *imgut.Image) float64 {
		// Images to vector
		dataInd := imgut.ToSlice(indImage)
		// Compute average pixels
		avgy := floats.Sum(dataInd) / float64(len(dataInd))
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
		return math.Sqrt(total / float64(len(dataInd)))
	}
}

func MakeFitEdge(targetImage *imgut.Image, stats map[string]*sequence.SequenceStats) func(*imgut.Image) float64 {
	// Compute edge detection
	edgeKern := &imgut.ConvolutionMatrix{3, []float64{
		0, 1, 0,
		1, -4, 1,
		0, 1, 0},
	}
	targEdge := imgut.ApplyConvolution(edgeKern, targetImage)
	// Function to compute RMSE
	rmseFit := MakeFitRMSE(targetImage)
	return func(indImg *imgut.Image) float64 {
		// Compute regular RMSE on this image
		rmse := rmseFit(indImg)

		imgEdge := imgut.ApplyConvolution(edgeKern, indImg)
		// Compute distance between edges
		edRmse := imgut.PixelRMSE(imgEdge, targEdge)

		// Statistics on output values
		if _, ok := stats["sub-fit-plain"]; !ok {
			stats["sub-fit-plain"] = sequence.Create()
		}
		stats["sub-fit-plain"].Observe(rmse)

		if _, ok := stats["sub-fit-edged"]; !ok {
			stats["sub-fit-edged"] = sequence.Create()
		}
		stats["sub-fit-edged"].Observe(edRmse)
		// Weighted fitness
		return rmse * edRmse
	}
}

func MakeFitSSIM(targetImage *imgut.Image) func(*imgut.Image) float64 {
	return func(indImage *imgut.Image) float64 {
		// Create temporary files
		tarTemp, _ := ioutil.TempFile(".", "tarImg")
		indTemp, _ := ioutil.TempFile(".", "indImg")
		// Schedule for cleaning
		defer os.Remove(tarTemp.Name())
		defer os.Remove(indTemp.Name())
		tarTemp.Close()
		indTemp.Close()
		// Save images to those files
		targetImage.WritePNG(tarTemp.Name())
		indImage.WritePNG(indTemp.Name())

		// Compute the DSSIM value (dissimilarity)
		cmd := exec.Command("./dssim.exe", tarTemp.Name(), indTemp.Name())
		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			print("ERRORE CON LA CALL", stderr.String())
			ooo, _ := cmd.CombinedOutput()
			print("Output ")
			print(ooo)
			panic(err)
		}
		res := strings.Split(out.String(), "\t")[0]
		fit, err := strconv.ParseFloat(res, 64)
		if err != nil {
			print("ERROR CON CONVERT")
			panic(err)
		}
		return fit
	}
}

// This method evaluates the current genotype and returns its fitness
// without caching the results (i.e. fitnessIsValid is NOT read or written)
func (ind *Individual) Evaluate() ga.Fitness {
	ind.set.Draw(ind, ind.ImgTemp)                  // Draw individual
	return ga.Fitness(ind.set.FitFunc(ind.ImgTemp)) // Evaluate fit
}

func (ind *Individual) FitnessValid() bool {
	return ind.fitIsValid
}

func (ind *Individual) Invalidate() {
	ind.fitIsValid = false
}

func (ind *Individual) Initialize() {
	ind.Node = ind.set.GenFunc(ind.set.MaxDepth)
	ind.ImgTemp = imgut.Create(ind.set.ImgTarget.W, ind.set.ImgTarget.H, ind.set.ImgTarget.ColorSpace)
}

func (ind *Individual) Mutate(pMut float64) {
	if ind.set.Mutate(pMut, ind) {
		ind.Invalidate()
	}
}

func (ind *Individual) CountEvent(name string, e bool) {
	// Statistics on output values
	if _, ok := ind.set.Counters[name]; !ok {
		ind.set.Counters[name] = new(counter.BoolCounter)
	}
	ind.set.Counters[name].Count(e)
}
