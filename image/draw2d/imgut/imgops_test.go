package imgut

import (
	"image/color"
	"math/rand"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())

	img := Create(100, 100, MODE_RGBA)
	img.SetColor(0.5, 0.5, 0.5)
	img.FillRect(0, 0, float64(img.W), float64(img.H), 0.5, 0.5, 0.5)
	path := "/home/akiross/Dropbox/Dottorato/TeslaPhD/GoGP/testing_image.png"
	img.WritePNG(path)
	t.Log("Loading the saved image...")
	img2, err := Load(path)
	t.Log("Is the image loaded?")
	if err != nil {
		t.Error("Cannot load the image we just saved!! WTF?!?!")
	}
	t.Log("Image loaded correctly :)")
	if PixelRMSE(img, img2) != 0 {
		t.Error("DISTANCE IS NOT NULL, WTF?!?!")
	}
	t.Log("Pixel distance is zero, ofc :)")
	img.FillRect(0, 0, float64(img.W), float64(img.H), 0.6, 0.6, 0.6)
	t.Log("Pixel distance now:", PixelRMSE(img, img2))
	img.FillTriangle(10, 50, 50, 1, 0, 0)
	img.FillTriangle(90, 50, 50, 0, 1, 0)
	img.WritePNG(path)

	//
	startCol, endCol := 0.0, 1.0
	sx, sy, ex, ey := 0.1, 0.5, 0.6, 0.5

	xd, yd := (ex - sx), (ey - sy)
	c1, c2 := xd*sx+yd*sy, xd*ex+yd*ey
	cd := c2 - c1

	img.FillMath(0, 0, 1, 1, func(x, y float64) float64 {
		c := xd*float64(x) + yd*float64(y)
		if c <= c1 {
			return startCol
		}
		if c >= c2 {
			return endCol
		}
		return (startCol*(c2-c) + endCol*(c-c1)) / cd
	})
	img.WritePNG(path)

	for i := 0; i < 10000; i++ {
		img.LinearShade(0, 0, 100, 100, rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64())
	}
	img.WritePNG(path)
}

func TestConvolution(t *testing.T) {
	t.Log("Testing convolution")
	cm3 := ConvolutionMatrix{3, []float64{
		0, 1, 0,
		1, -4, 1,
		0, 1, 0},
	}
	//cm3.Normalize()

	cm5 := ConvolutionMatrix{5, []float64{
		0, 0, 0, 0, 0,
		0, 0, 1, 0, 0,
		0, 1, -4, 1, 0,
		0, 0, 1, 0, 0,
		0, 0, 0, 0, 0},
	}
	cm5.Normalize()

	path := "lena.png"
	t.Log("Loading sample image...")
	img, err := Load(path)
	if err != nil {
		t.Error("Cannot load the image we just saved!! WTF?!?!")
	}

	img3 := ApplyConvolution(&cm3, img)
	img3.WritePNG("/home/akiross/Dropbox/Dottorato/TeslaPhD/GoGP/test_conv3.png")
	//	img5 := ApplyConvolution(&cm5, img)
	//	img5.WritePNG("/home/akiross/Dropbox/Dottorato/TeslaPhD/GoGP/test_conv5.png")
}

func TestComposition(t *testing.T) {
	img1 := Create(100, 100, MODE_RGBA)
	img2 := Create(100, 100, MODE_RGBA)
	img3 := Create(100, 100, MODE_RGBA)

	img2.FillMath(0, 0, 1, 1, func(x, y float64) float64 { return y })
	img3.FillMath(0, 0, 1, 1, func(x, y float64) float64 { return x })
	img1.FillVectorialize([]*Image{img2, img3}, func(comps []color.Color) color.Color {
		cm1 := img1.Surf.ColorModel()
		var tot float64
		for i := range comps {
			c := cm1.Convert(comps[i]).(color.RGBA)
			tot += float64(c.R) / 255.0
		}
		if tot < 0 {
			tot = 0
		} else if tot > 1 {
			tot = 1
		}
		return color.RGBA{uint8(tot * 0xff), 0x00, 0x00, 0xff}
	})
	img1.WritePNG("img1.png")
	img2.WritePNG("img2.png")
	img3.WritePNG("img3.png")
}

// Bah, not working
/*
func TestSobel(t *testing.T) {
	img, err := Load("lena.png")
	if err != nil {
		t.Error("Cannot load the image we just saved!! WTF?!?!")
	}
	sob := SobelOperator(img)
	sob.WritePNG("lena_sobel.png")
}
*/
