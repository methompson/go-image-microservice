package imageConversion

import (
	"image"
	"math"

	"github.com/nfnt/resize"
)

func scaleImage(img *image.Image, longestSide uint) *image.Image {
	width := float64((*img).Bounds().Max.X)
	height := float64((*img).Bounds().Max.Y)

	longestFloat := float64(longestSide)
	newShorter := calculateShorterDimension(width, height, longestFloat)

	var newWidth uint
	var newHeight uint
	// If width is greater than height
	if width > height {
		newWidth = longestSide
		newHeight = newShorter
	} else {
		newHeight = longestSide
		newWidth = newShorter
	}

	var image = resize.Resize(newWidth, uint(newHeight), *img, resize.Lanczos3)

	return &image
}

// For this, we calculate the AR, longer / shorter, then divide the longest
// side by the AR to get the other side length. We round to even for better
// image compression and convert to an integer
func calculateShorterDimension(side1, side2, newLongSide float64) uint {
	var longerSide float64
	var shorterSide float64

	if side1 > side2 {
		longerSide = side1
		shorterSide = side2
	} else {
		longerSide = side2
		shorterSide = side1
	}

	return uint(math.RoundToEven(newLongSide / (longerSide / shorterSide)))
}

func makeThumbnail(img *image.Image) *image.Image {
	var thumb = resize.Thumbnail(128, 128, *img, resize.Lanczos3)
	return &thumb
}
