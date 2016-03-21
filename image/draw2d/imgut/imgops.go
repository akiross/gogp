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
	"sync"
	"unsafe"
)

// BUG(akiross) Basically, only RGBA is supported (and has been shallowly tested)

// #cgo CFLAGS: -O2 -Wall -fopenmp
// #cgo LDFLAGS: -lgomp
// #include "linearShading.h"
// #include "imageAverage.h"
import "C"

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
	img.Surf = image.NewRGBA(image.Rect(0, 0, w, h))
	img.Ctx = draw2dimg.NewGraphicContext(img.Surf)
	img.W, img.H = w, h
	return &img
}

func getDataPointer(img *Image) (*C.uchar, C.int) {
	rgbaImage := img.Surf.(*image.RGBA)
	return (*C.uchar)(unsafe.Pointer(&rgbaImage.Pix[0])), C.int(rgbaImage.Stride)
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

func (img *Image) LinearShade(x1, y1, x2, y2, sx, sy, ex, ey, startCol, endCol float64) {
	// Get a pointer to image data
	pixPtr, stride := getDataPointer(img)
	//rgbaImage := img.Surf.(*image.RGBA)
	//pixPtr := (*C.uchar)(unsafe.Pointer(&rgbaImage.Pix[0]))
	// Do shading
	C.linearShading(pixPtr, stride, C.int(x1), C.int(y1), C.int(x2), C.int(y2), C.double(startCol), C.double(endCol), C.double(sx), C.double(sy), C.double(ex), C.double(ey))
}

func (img *Image) CircularShade(cx, cy, inRad, outRad, startCol, endCol float64) {
	//	pixPtr, stride := getDataPointer(img)
	//	C.circularShading(pixPtr, stride, C.int(cx), C.int(cy))
}

// Copy image onto the target image, at specified position
func (i *Image) Blit(x, y int, target *Image) {
	destPoint := image.Pt(x, y)
	destRect := image.Rectangle{destPoint, destPoint.Add(image.Pt(i.W, i.H))}
	draw.Draw(target.Surf, destRect, i.Surf, image.ZP, draw.Src)
}

// Clear the image filling with black
func (i *Image) Clear() {
	draw.Draw(i.Surf, i.Surf.Bounds(), image.Transparent, image.ZP, draw.Src)
}

type PixelFunc func(x, y float64) float64

// Fill the image evaluating the function over each pixel
// You can specify one function per channel, or one function for all the channels
// The values will be normalized
// BUG(akiross) this should return an error
// Coordinates are relative to image boundary sizes
func (img *Image) FillMath(minX, minY, maxX, maxY float64, chanFuncs ...PixelFunc) {
	b := img.Surf.Bounds()
	bx, by := int(float64(b.Min.X)*minX), int(float64(b.Min.Y)*minY)
	ex, ey := int(float64(b.Max.X)*maxX), int(float64(b.Max.Y)*maxY)

	// Wait many goroutines to finish
	var wg sync.WaitGroup

	// Number of goroutines to wait for
	wg.Add(ey - by)

	// Start goroutines
	for i := by; i < ey; i++ {
		go func(i int) {
			for j := bx; j < ex; j++ {
				val := uint8(chanFuncs[0](float64(i)/float64(ey-by), float64(j)/float64(ex-bx)) * 0xff)
				col := color.RGBA{val, val, val, 0xff}
				img.Surf.Set(i, j, col)
			}
			wg.Done()
		}(i)
	}

	// Wait for them to finish
	wg.Wait()
}

func (img *Image) FillMathBounds(chanFuncs ...PixelFunc) {
	img.FillMath(0, 0, 1, 1, chanFuncs...)
}

// Fill img pixel-by-pixel using the result of the comp function called on every pixel
// of every image in images. images must have the same boundaries and color space of img
func (img *Image) FillVectorialize(images []*Image, comp func([]color.Color) color.Color) {
	b := img.Surf.Bounds()
	for i := b.Min.Y; i < b.Max.Y; i++ {
		for j := b.Min.X; j < b.Max.X; j++ {
			colors := make([]color.Color, len(images))
			for k := range colors {
				colors[k] = images[k].Surf.At(j, i)
			}
			img.Surf.Set(j, i, comp(colors))
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

/*
// Compute the distance between two images, pixel by pixel (RSME)
func PixelDistance(i1, i2 *Image) (rmse float64) {
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
	// BUG(akiross) in the python version this is not normalized
	rmse = math.Sqrt(rmse) / float64(count)
	return
}
*/

func clamp8(v float64) uint8 {
	if int(v) < 0 {
		return 0
	} else if int(v) > 255 {
		return 255
	}
	return uint8(v)
}

func ToSlice(img *Image) []float64 {
	surf := img.Surf
	cm := surf.ColorModel()
	b := surf.Bounds()
	minX, maxX := b.Min.X, b.Max.X
	minY, maxY := b.Min.Y, b.Max.Y
	data := make([]float64, (maxX-minX)*(maxY-minY)*4)
	k := 0
	for i := b.Min.Y; i < b.Max.Y; i++ {
		for j := b.Min.X; j < b.Max.X; j++ {
			pix := surf.At(j, i)
			col := cm.Convert(pix).(color.RGBA)
			data[k+0] = float64(col.R)
			data[k+1] = float64(col.G)
			data[k+2] = float64(col.B)
			data[k+3] = float64(col.A)
			k += 4
		}
	}
	return data
}

func FromSlice(img *Image, data []float64) {
	surf := img.Surf
	cm := surf.ColorModel()
	b := surf.Bounds()
	minX, maxX := b.Min.X, b.Max.X
	minY, maxY := b.Min.Y, b.Max.Y
	k := 0
	for i := minY; i < maxY; i++ {
		for j := minX; j < maxX; j++ {
			const mk = 0xff
			outCol := color.RGBA{
				clamp8(data[k+0]) & mk,
				clamp8(data[k+1]) & mk,
				clamp8(data[k+2]) & mk,
				clamp8(data[k+3]) & mk,
			}
			k += 4
			img.Surf.Set(j, i, cm.Convert(outCol))
		}
	}
}

// Compute the distance between two images, pixel by pixel (RSME)
func PixelRMSE(i1, i2 *Image) (rmse float64) {
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
	rmse = math.Sqrt(rmse / float64(count))
	return
}

// Returns the average image of the provided slice of images
// The output size will be the same as first input image.
// Any image with different size will be ignored.
func Average(images []*Image) *Image {
	if len(images) == 0 {
		return nil
	}

	width, height := images[0].W, images[0].H

	// Build storage for accumulated image
	accumulator := make([]float32, width*height*4)
	accPtr := (*C.float)(unsafe.Pointer(&accumulator[0]))

	// Accumulate every image
	for i := range images {
		// Pixel-by-pixel
		pixPtr, stride := getDataPointer(images[i])
		//		rgbaImage := images[i].Surf.(*image.RGBA)
		//		pixPtr := (*C.uchar)(unsafe.Pointer(&rgbaImage.Pix[0]))
		C.imageAccumulate(accPtr, pixPtr, stride, C.int(width), C.int(height))
	}

	// Output image
	avgImg := Create(width, height, images[0].ColorSpace)
	imgPtr, stride := getDataPointer(avgImg)
	// The resulting image is the accumulator divided by number of images
	C.imageDivide(imgPtr, accPtr, C.float(len(images)), stride, C.int(width), C.int(height))

	return avgImg
}

type ConvolutionMatrix struct {
	Size int
	Data []float64
}

// Normalize the matrix so that Data sums to 1
func (cm *ConvolutionMatrix) Normalize() {
	var tot float64
	for i := range cm.Data {
		tot += cm.Data[i]
	}
	for i := range cm.Data {
		cm.Data[i] /= tot
	}
}

// WARNING: This function does NOT apply the convolution filter on alpha
// and uses the original alpha value for each img pixel
func (cm *ConvolutionMatrix) Multiply(img *Image, x, y int) color.RGBA {
	var racc, gacc, bacc float64
	count := cm.Size * cm.Size
	b := img.Surf.Bounds()

	for k := 0; k < count; k++ {
		i, j := k/cm.Size-cm.Size/2, k%cm.Size-cm.Size/2
		y2, x2 := y+i, x+j
		if y2 < b.Min.Y || y2 >= b.Max.Y {
			y2 = y
		}
		if x2 < b.Min.X || x2 >= b.Max.X {
			x2 = x
		}

		rgba8 := color.RGBAModel.Convert(img.Surf.At(x2, y2)).(color.RGBA)

		nr := float64(rgba8.R) * cm.Data[k]
		ng := float64(rgba8.G) * cm.Data[k]
		nb := float64(rgba8.B) * cm.Data[k]

		racc += nr
		gacc += ng
		bacc += nb
	}

	const mk = 0xff
	outCol := color.RGBA{
		clamp8(racc) & mk,
		clamp8(gacc) & mk,
		clamp8(bacc) & mk,
		img.Surf.At(x, y).(color.RGBA).A, // XXX original alpha value
	}
	return outCol
}

func ApplyConvolution(cm *ConvolutionMatrix, img *Image) *Image {
	// Create result image
	dest := Create(img.W, img.H, img.ColorSpace)

	bs := img.Surf.Bounds()
	for i := bs.Min.Y; i < bs.Max.Y; i++ {
		for j := bs.Min.X; j < bs.Max.X; j++ {
			dest.Surf.Set(j, i, cm.Multiply(img, j, i))
		}
	}

	return dest
}

/* Not a good one... FIXME
// Returns a new image with the sobel operator applied to img
func SobelOperator(img *Image) *Image {
	gx := &ConvolutionMatrix{3, []float64{
		-1, 0, 1,
		-2, 0, 2,
		-1, 0, 1},
	}
	gy := &ConvolutionMatrix{3, []float64{
		-1, -2, -1,
		0, 0, 0,
		1, 2, 1},
	}

	imgs := make([]*Image, 2)
	imgs[0] = ApplyConvolution(gx, img)
	imgs[1] = ApplyConvolution(gy, img)

	out := Create(img.W, img.H, img.ColorSpace)

	out.FillVectorialize(imgs, func(comps []color.Color) color.Color {
		cm := img.Surf.ColorModel()

		c1 := cm.Convert(comps[0]).(color.RGBA)
		c2 := cm.Convert(comps[1]).(color.RGBA)

		const maxVal = 360.7 // (0xff * 0xff + 0xff * 0xff) ** 0.5

		r1, g1, b1 := float64(c1.R), float64(c1.G), float64(c1.B)
		r2, g2, b2 := float64(c2.R), float64(c2.G), float64(c2.B)

		r := clamp8(0xff * math.Sqrt(r1*r1+r2+r2) / maxVal)
		g := clamp8(0xff * math.Sqrt(g1*g1+g2+g2) / maxVal)
		b := clamp8(0xff * math.Sqrt(b1*b1+b2+b2) / maxVal)
		return color.RGBA{r, g, b, 0xff}
	})
	return out
}
*/
