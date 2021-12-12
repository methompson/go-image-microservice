package imageHandler

import (
	"image"
	"math"
	"os"
	"strconv"

	"github.com/nfnt/resize"
	"methompson.com/image-microservice/imageServer/constants"
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

	var image = resize.Resize(newWidth, newHeight, *img, resize.Lanczos3)

	return &image
}

func scaleImageByWidth(img *image.Image, newWidth uint, rotated bool) *image.Image {
	if rotated {
		return scaleImageByY(img, newWidth)
	} else {
		return scaleImageByX(img, newWidth)
	}
}

// Scales an image such that the Aspect ratio is (mostly) constrained. Scales the
// X value so that it aligns with the newY value.
func scaleImageByY(img *image.Image, newY uint) *image.Image {
	X := float64((*img).Bounds().Max.X)
	Y := float64((*img).Bounds().Max.Y)

	newX := calculateOtherSide(Y, X, float64(newY))

	var image = resize.Resize(newX, newY, *img, resize.Lanczos3)

	return &image
}

// Scales an image such that the Aspect ratio is (mostly) constrained. Scales the
// Y value so that it aligns with the newX value.
func scaleImageByX(img *image.Image, newX uint) *image.Image {
	X := float64((*img).Bounds().Max.X)
	Y := float64((*img).Bounds().Max.Y)

	newY := calculateOtherSide(X, Y, float64(newX))

	var image = resize.Resize(newX, newY, *img, resize.Lanczos3)
	// var image = resize.Resize(newY, newX, *img, resize.Lanczos3)

	return &image
}

// Given two sides, side1 and side2, this function calculates new side 2 when given
// new side 1 by finding the aspect ratio between the two sides. The end result is
// a new side the produces the same or similar aspect ratio
func calculateOtherSide(side1, side2, newSide1 float64) uint {
	return uint(math.RoundToEven(newSide1 / (side1 / side2)))
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

	return calculateOtherSide(longerSide, shorterSide, newLongSide)
}

// Scales an image using the resize.Thumbnail function.
func makeThumbnail(img *image.Image) *image.Image {
	dim := getThumbnailDimenions()
	var thumb = resize.Thumbnail(dim, dim, *img, resize.Lanczos3)
	return &thumb
}

// Gets jpeg quality as an integer. Retrieves the value from the env
// and if it doesn't exist or the value is erroneous, returns 75 as
// a default
func getJpegQuality() int {
	val, err := strconv.Atoi(os.Getenv(constants.JPEG_QUALITY))

	if err != nil || val < 1 || val > 100 {
		return 75
	}

	return val
}

// Gets dimensions for a thumbnail. Retrieves the value from the env
// and if it doesn't exist or the value is erroneous, returns 128 as
// a default
func getThumbnailDimenions() uint {
	val, err := strconv.ParseUint(os.Getenv(constants.THUMBNAIL_SIZE), 10, 0)

	if err != nil || val > 100 {
		return 128
	}

	return uint(val)
}
