package imageConversion

import (
	"bytes"
	"encoding/base64"
	"image"
	"testing"

	"github.com/nfnt/resize"
)

func TestCalculateShorterDimension(t *testing.T) {
	var result uint

	result = calculateShorterDimension(1024, 768, 640)

	if result != 480 {
		t.Fatalf("result = '%v', Should be '480'", result)
	}

	result = calculateShorterDimension(768, 1024, 640)

	if result != 480 {
		t.Fatalf("result = '%v', Should be '480'", result)
	}

	result = calculateShorterDimension(1024, 1024, 640)

	if result != 640 {
		t.Fatalf("result = '%v', Should be '640'", result)
	}
}

func TestScaleImage(t *testing.T) {
	// PNG files in Base 64
	oneByOneb64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+P+/HgAFhAJ/wlseKgAAAABJRU5ErkJggg=="
	oneByOneBytes, decodeErr := base64.StdEncoding.DecodeString(oneByOneb64)

	if decodeErr != nil {
		t.Fatalf("Error decode base64 string")
	}

	oneBeOneImage, _, decodeErr := image.Decode(bytes.NewReader(oneByOneBytes))

	if decodeErr != nil {
		t.Fatalf("Error decoding image")
	}

	_1024x768 := resize.Resize(1024, 768, oneBeOneImage, resize.Lanczos3)

	_640x480 := scaleImage(&_1024x768, 640)

	width := (*_640x480).Bounds().Max.X
	height := (*_640x480).Bounds().Max.Y

	if width != 640 {
		t.Fatalf("width is '%v'. Should be '640", width)
	}
	if height != 480 {
		t.Fatalf("height is '%v'. Should be '480", height)
	}

	_768x1024 := resize.Resize(768, 1024, oneBeOneImage, resize.Lanczos3)

	_480x640 := scaleImage(&_768x1024, 640)

	width = (*_480x640).Bounds().Max.X
	height = (*_480x640).Bounds().Max.Y

	if width != 480 {
		t.Fatalf("width is '%v'. Should be '480", width)
	}
	if height != 640 {
		t.Fatalf("height is '%v'. Should be '640", height)
	}
}
