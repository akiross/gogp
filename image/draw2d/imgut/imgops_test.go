package imgut

import (
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
	if PixelDistance(img, img2) != 0 {
		t.Error("DISTANCE IS NOT NULL, WTF?!?!")
	}
	t.Log("Pixel distance is zero, ofc :)")
	img.FillRect(0, 0, float64(img.W), float64(img.H), 0.6, 0.6, 0.6)
	t.Log("Pixel distance now:", PixelDistance(img, img2))
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
