package base

import (
	"github.com/akiross/gogp/image/draw2d/imgut"
	"math/rand"
	"testing"
	"time"
)

func TestRMSE(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	// Create a white image
	img1 := imgut.Create(10, 1, imgut.MODE_RGBA)
	img1.FillRect(0, 0, 10, 1, 1.0, 1.0, 1.0)
	// Create a black image
	img2 := imgut.Create(10, 1, imgut.MODE_RGBA)
	img2.FillRect(0, 0, 10, 1, 0.0, 0.0, 0.0)
	//path := "/home/akiross/Dropbox/Dottorato/TeslaPhD/GoGP/rmse_test_image.png"
	//img2.WritePNG(path)

	// Compare RMSEs
	rmseVec := fitnessRMSE(img1, img2)
	rmseImg := fitnessRMSEImage(img1, img2)
	t.Log("Vec RMSE", rmseVec, rmseImg)
}
