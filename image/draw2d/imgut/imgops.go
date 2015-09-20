package imgut

import (
	"fmt"
	"github.com/llgcode/draw2d/draw2dimg"
	"image"
	_ "image/png"
	"math"
	"os"
	"unsafe"
)

type ColorSpace int

const (
	MODE_A8 ColorSpace = iota
	MODE_G8
	MODE_RGB
	MODE_RGBA
)

type Image struct {
	Surf       image.Image              // Image data
	Ctx        draw2dimg.GraphicContext // Drawing context
	W, H       int                      // Image size
	ColorSpace ColorSpace               // What colors are we considering
}

// Create an image of the given size
func Create(w, h int, mode ColorSpace) *Image {
	var img Image
	switch mode {
	default:
		img.Surf = image.NewGray(image.Rect(0, 0, w, h))
	case MODE_RGB, MODE_RGBA:
		img.Surf = image.NewRGBA(image.Rect(0, 0, w, h))
	}
	img.ColorSpace = mode
	img.Ctx = draw2dimg.NewGraphicsContext(img.Surf)
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

	img.Surf, err = png.Decode(file)
	if err != nil {
		fmt.Println("ERROR: Cannot decode PNG file", path)
		return nil, err
	}

	img.Ctx = draw2dimg.NewGraphicsContext(img.Surf)
	br := img.Surf.Bounds()
	img.W, img.H = br.Max.X-br.Min.X, br.Max.Y-br.Min.Y

	fmt.Println("Loading image with model", img.Surf.ColorModel())

	return &img
}

func (i *Image) SetColor(col ...float64) {
	var r, g, b, a byte = 0, 0, 0, 0xff
	// Single channel, use alpha
	if len(col) == 1 {
		if i.ColorSpace == MODE_A8 {
			// When using alphas, draw on alpha
			a = byte(col[0] * 0xff)
		} else if i.ColorSpace == MODE_G8 {
			// When using grayscale, draw RGB
			r = byte(col[0] * 0xff)
			g = r
			b = r
		}
	} else if len(col) >= 3 {
		r, g, b = byte(col[0]*0xff), byte(col[1]*0xff), byte(col[2]*0xff)
	}
	if len(col) == 4 {
		a = C.double(col[3])
	}
	i.Ctx.SetFillColor(color.RGBA{r, g, b, a})
}

// Stroke the current path with the given color
func (i *Image) StrokeColor(col ...float64) {
	i.SetColor(col...)
	C.cairo_stroke(i.Ctx)
}

// Fill the current path with the given color
func (i *Image) FillColor(col ...float64) {
	i.SetColor(col...)
	C.cairo_fill(i.Ctx)
}

// Draw given poligon path, automatically closing first and last point
func (i *Image) DrawPoly(points ...float64) {
	C.cairo_move_to(i.Ctx, C.double(points[0]), C.double(points[1]))
	for j := 2; j < len(points); j += 2 {
		C.cairo_line_to(i.Ctx, C.double(points[j]), C.double(points[j+1]))
	}
	C.cairo_close_path(i.Ctx)
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
	nch := len(chanFuncs)
	// BUG(akiross) this code introduces checks which may be unnecessary
	// in production, would be nice to disable this checking
	switch img.ColorSpace {
	case MODE_A8:
		if nch != 1 {
			fmt.Println("ERROR! For A8 images is mandatory to use a single chanFunc")
			return
		}
	case MODE_G8:
		if nch != 1 {
			fmt.Println("ERROR! For G8 images is mandatory to use a single chanFunc")
			return
		}
	case MODE_RGB:
		if nch < 1 || nch > 3 {
			fmt.Println("ERROR! For mode RGB you need 1, 2 or 3 chanFuncs")
			return
		}
	default:
		fmt.Println("RGBA not implemented yet! Alpha must evaluate to 1, or value is always transparent!")
		return
		/*
			if nch < 1 || nch > 4 {
				fmt.Println("ERROR! For mode RGBA you need from 1 to 4 chanFuncs")
				return
			}
			if nch == 4 {
				// If 4 parameters are specified, the alpha gets moved in front
				chanFuncs[0], chanFuncs[1], chanFuncs[2], chanFuncs[3] = chanFuncs[3], chanFuncs[0], chanFuncs[1], chanFuncs[2]
			} else {
				// Else, first parameter is alpha (nil), other parameters follow
				chanFuncs2 := []PixelFunc{nil}
				chanFuncs = append(chanFuncs2, chanFuncs...)
			}
				Problem: if alpha returns a constant value, or if it is nil, image is always transparent (normalized -> 0)
		*/
	}
	// Evaluate the channel functions in every point
	realData := make([][][]float64, nch)
	//realData := make([]float64, img.H*img.W*nch)
	max := make([]float64, nch) // Hold max value per each channel
	min := make([]float64, nch) // Hold min value per each channel
	for k := 0; k < nch; k++ {
		realData[k] = make([][]float64, img.H)
		// If the channel function is defined, use it on every pixel
		// othersize, the data, min and max are already set to 0
		if chanFuncs[k] != nil {
			for i := 0; i < img.H; i++ {
				realData[k][i] = make([]float64, img.W)
				for j := 0; j < img.W; j++ {
					realData[k][i][j] = chanFuncs[k](j, i)
					max[k] = math.Max(max[k], realData[k][i][j])
					min[k] = math.Min(min[k], realData[k][i][j])
				}
			}
		}
	}

	// Copy the data onto the image
	stride := int(C.cairo_image_surface_get_stride(img.Surf)) // Stride in bytes
	rawData := unsafe.Pointer(C.cairo_image_surface_get_data(img.Surf))

	// Prepare byte data, normalizing if necessary (we cannot write directly to unsafe.Pointer)
	byteData := make([]byte, stride*img.H)

	// Depending on format, we copy the data in different ways
	switch img.ColorSpace {
	case MODE_A8:
		fmt.Println("Copying data for mode A8")
		const k = 0
		if max[k] != min[k] {
			for i := 0; i < img.H; i++ {
				for j := 0; j < img.W; j++ {
					byteData[i*stride+j] = byte(0xff * (realData[k][i][j] - min[k]) / (max[k] - min[k]))
				}
			}
		}
	case MODE_G8:
		fmt.Println("Copying data for mode G8")
		const k = 0
		if max[k] != min[k] {
			for i := 0; i < img.H; i++ {
				for j := 0; j < img.W; j++ {
					p := i*stride + j*4
					v := byte(0xff * (realData[k][i][j] - min[k]) / (max[k] - min[k]))
					byteData[p], byteData[p+1], byteData[p+2] = v, v, v
				}
			}
		}
	case MODE_RGB:
		fmt.Println("Copying data mode RGB")
		for k := 0; k < nch; k++ {
			if max[k] != min[k] {
				for i := 0; i < img.H; i++ {
					for j := 0; j < img.W; j++ {
						p := i*stride + j*4 + 1
						byteData[p+k] = byte(0xff * (realData[k][i][j] - min[k]) / (max[k] - min[k]))
					}
				}
			}
		}
	/*
		case MODE_RGBA:
			fmt.Println("Copying data mode RGBA")
			for k := 0; k < nch; k++ {
				if max[k] != min[k] {
					for i := 0; i < img.H; i++ {
						for j := 0; j < img.W; j++ {
							p := i*stride + j*4 + 1
							byteData[p+k] = byte(0xff * (realData[k][i][j] - min[k]) / (max[k] - min[k]))
						}
					}
				}
			}
	*/
	default:
		fmt.Println("ERROR! Not implemented yet")
	}
	// Copy the data on the C-side
	C.memcpy(rawData, unsafe.Pointer(&byteData[0]), C.size_t(stride*img.H))
}

// Write an image to PNG
func (i *Image) WritePNG(file string) {
	C.cairo_surface_write_to_png(i.Surf, C.CString(file))
}

// Get the data of the image for the specified channels
// BUG(akiross) this should return an error, instead of printing!
func (img *Image) GetChannels(ch ...int) [][][]byte {
	// Number of requested channels. If none is specified, all are taken
	nch := len(ch)
	// Number of existing channels
	var ech int

	// Verify that format is compatible with the channel request
	if true {
		switch img.ColorSpace {
		case MODE_A8:
			ech = 1
			if nch == 0 {
				nch = 1
			} else if nch != 1 {
				fmt.Println("ERROR! For MODE_A8 you must require one channel")
				return nil
			}
		case MODE_G8:
			ech = 3
			if nch == 0 {
				nch = 3
			} else if nch != 1 && nch != 3 {
				fmt.Println("ERROR! For MODE_G8, you may pick 1 or 3 chans!")
				return nil
			}
		case MODE_RGB:
			ech = 3
			if nch == 0 {
				nch = 3
			} else if nch < 1 || nch > 3 {
				fmt.Println("ERROR! For MODE_RGB, you may pick 1, 2 or 3 chans")
				return nil
			}
		case MODE_RGBA:
			ech = 4
			if nch == 0 {
				nch = 4
			} else if nch < 1 || nch > 4 {
				fmt.Println("ERROR! For MODE_RGBA, you may pick 1, 2, 3 or 4 chans")
				return nil
			}
		default:
			fmt.Println("ERROR! Unsupported color space", img.ColorSpace)
			return nil
		}
	}

	// Get the raw data
	stride := int(C.cairo_image_surface_get_stride(img.Surf))
	rawData := C.GoBytes(unsafe.Pointer(C.cairo_image_surface_get_data(img.Surf)), C.int(stride*img.H))

	// Build the matrix
	mtx := make([][][]byte, img.H)
	for i := 0; i < img.H; i++ {
		mtx[i] = make([][]byte, img.W)
		for j := 0; j < img.W; j++ {
			mtx[i][j] = make([]byte, nch)
			for k := 0; k < nch; k++ {
				mtx[i][j][k] = rawData[i*stride+j*ech+k]
			}
		}
	}
	return mtx
}
