package imgut

import (
	"fmt"
	"github.com/llgcode/draw2d/draw2dimg"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	//	"unsafe"
)

type ColorSpace int

const (
	MODE_A8 ColorSpace = iota
	MODE_G8
	MODE_RGB
	MODE_RGBA
)

type Image struct {
	Surf       draw.Image                // Image data
	Ctx        *draw2dimg.GraphicContext // Drawing context
	W, H       int                       // Image size
	ColorSpace ColorSpace                // What colors are we considering
}

// Create an image of the given size
func Create(w, h int, mode ColorSpace) *Image {
	var img Image
	// BUG(akiross) only RGBA supported
	img.Surf = image.NewRGBA(image.Rect(0, 0, w, h))
	img.Ctx = draw2dimg.NewGraphicContext(img.Surf)
	img.W, img.H = w, h
	return &img
}

// Load an image from a PNG file
func Load(path string) (*Image, error) {
	var img Image

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("ERROR: Cannot open file", path)
		return nil, err
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	pngImage, err := png.Decode(file)
	if err != nil {
		fmt.Println("ERROR: Cannot decode PNG file", path)
		return nil, err
	}

	// Copy to surface
	b := pngImage.Bounds()
	img.Surf = image.NewRGBA(image.Rect(0, 0, b.Max.X-b.Min.X, b.Max.Y-b.Min.Y))
	draw.Draw(img.Surf, img.Surf.Bounds(), pngImage, pngImage.Bounds().Min, draw.Src)

	img.Ctx = draw2dimg.NewGraphicContext(img.Surf)
	br := img.Surf.Bounds()
	img.W, img.H = br.Max.X-br.Min.X, br.Max.Y-br.Min.Y

	return &img, nil
}

func (i *Image) SetColor(col ...float64) {
	// BUG(akiross) RGBA is supported right now
	if len(col) < 3 {
		fmt.Println("ERROR: SetColor works only in RGB mode right now, you need to pass 3 parameters")
		panic("ERROR: SetColor requires 3 parameters")
	}
	i.Ctx.SetFillColor(color.RGBA{uint8(col[0] * 0xff), uint8(col[1] * 0xff), uint8(col[2] * 0xff), 0xff})
}

// Stroke the current path with the given color
func (i *Image) StrokeColor(col ...float64) {
	i.SetColor(col...)
	i.Ctx.Stroke()
}

// Fill the current path with the given color
func (i *Image) FillColor(col ...float64) {
	i.SetColor(col...)
	i.Ctx.Fill()
}

// Draw given poligon path, automatically closing first and last point
func (i *Image) DrawPoly(points ...float64) {
	i.Ctx.MoveTo(points[0], points[1])
	for j := 2; j < len(points); j += 2 {
		i.Ctx.LineTo(points[j], points[j+1])
	}
	i.Ctx.Close()
}

func (i *Image) FillRect(x1, y1, x2, y2 float64, col ...float64) {
	i.DrawPoly(x1, y1, x2, y1, x2, y2, x1, y2)
	i.FillColor(col...)
}

func TriangleCenterY(x1, x2, y float64) float64 {
	const sin_60 = 0.86602540378443864676
	return y + (x1-x2)*sin_60
}

func (i *Image) DrawTriangle(x1, x2, y float64) {
	cy := TriangleCenterY(x1, x2, y)
	i.DrawPoly(x1, y, 0.5*(x1+x2), cy, x2, y)
}

func (i *Image) FillTriangle(x1, x2, y float64, col ...float64) {
	i.DrawTriangle(x1, x2, y)
	i.FillColor(col...)
}

func (i *Image) FillSurface(col ...float64) {
	i.FillRect(0, 0, float64(i.W), float64(i.H), col...)
}

type PixelFunc func(x, y int) float64

// Fill the image evaluating the function over each pixel
// You can specify one function per channel, or one function for all the channels
// The values will be normalized
// BUG(akiross) this should return an error
func (img *Image) FillMath(chanFuncs ...PixelFunc) {
	// BUG(akiross) Only RGBA is supported
	b := img.Surf.Bounds()
	for i := b.Min.Y; i < b.Max.Y; i++ {
		for j := b.Min.X; j < b.Max.X; j++ {
			val := uint8(chanFuncs[0](i, j) * 0xff)
			col := color.RGBA{val, val, val, 0xff}
			img.Surf.Set(i, j, col)
		}
	}
}

// Write an image to PNG
func (i *Image) WritePNG(path string) {
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("ERROR: Cannot open file for writing", path)
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	err = png.Encode(file, i.Surf)
	if err != nil {
		fmt.Println("ERROR: Cannot encode PNG file", path)
		return
	}
}

// Compute the distance between two images, pixel by pixel (RSME)
func PixelDistance(i1, i2 *Image) (rmse float64) {
	// BUG(akiross) Only RGBA is supported
	// Check that sizes are the same
	im1, im2 := i1.Surf, i2.Surf
	if im1.Bounds() != im2.Bounds() {
		fmt.Println("ERROR! Cannot compute distances for different sizes", im1.Bounds(), im2.Bounds())
		return
	}
	// Compute the distance
	var count int
	rmse = 0
	b := im1.Bounds()
	cm1, cm2 := im1.ColorModel(), im2.ColorModel()
	for i := b.Min.Y; i < b.Max.Y; i++ {
		for j := b.Min.X; j < b.Max.X; j++ {
			// In the python version this is not normalized, but I think it should be for precision issues
			px1, px2 := im1.At(j, i), im2.At(j, i)
			gr1, gr2 := cm1.Convert(px1).(color.RGBA), cm2.Convert(px2).(color.RGBA)
			diff := (float64(gr1.R) - float64(gr2.R)) // / 255.0
			rmse += diff * diff
			count++
		}
	}
	// BUG(akiross) in the python version this is not normalized, but it should be!! For now, use un-normalized version to compare results, later fix this bug
	// ind.fitness = gogp.Fitness(math.Sqrt(dist / float64(count)))
	rmse = math.Sqrt(rmse)
	return
}
