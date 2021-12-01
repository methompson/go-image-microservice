package imageConversion

import (
	"bytes"
	"errors"
	"image"
	"io"
	"strings"

	gif "image/gif"
	jpeg "image/jpeg"
	png "image/png"

	bmp "golang.org/x/image/bmp"
	tiff "golang.org/x/image/tiff"
)

type ImageType int8

const (
	Jpeg ImageType = iota
	Png
	Gif
	Bmp
	Tiff
)

type ImageSize struct {
	Width  int
	Height int
}

type ImageSizeFormat struct {
	Filename  string
	ImageSize *ImageSize
	FileSize  int
}

func MakeImageSizeFormat(filename string, fileSize int, imageSize *ImageSize) *ImageSizeFormat {
	return &ImageSizeFormat{
		Filename:  filename,
		ImageSize: imageSize,
		FileSize:  fileSize,
	}
}

type ResizeOp int8

const (
	Original ResizeOp = iota
	Thumbnail
	Scale
	ScaleByWidth
)

// The ConversionOp is a blueprint for an image conversion operation.
// We use the original source image and perform operations on that.
// The ConversionOp also defines a format to encode to, in case the
// user wants to have different sizes in different formats.
type ConversionOp struct {
	Suffix   string
	EncodeTo *ImageType

	// Resize Op information
	LongestSide uint
	ResizeOp    ResizeOp

	// Flags
	Obfuscate bool // Whether to keep all filenames the same or randomize the names
}

func makeOpFromRequest(req *ConversionRequest) *ConversionOp {
	var resizeOp ResizeOp
	switch strings.ToLower(req.ResizeOp) {
	case "thumbnail":
		resizeOp = Thumbnail
	case "thumb":
		resizeOp = Thumbnail
	case "scale":
		resizeOp = Scale
	case "scalebywidth":
		resizeOp = ScaleByWidth
	default:
		resizeOp = Original
	}

	var encodeTo *ImageType
	if req.CompressTo != nil {
		switch strings.ToLower(*req.CompressTo) {
		case "jpeg":
			tmp := Jpeg
			encodeTo = &tmp
		case "png":
			tmp := Png
			encodeTo = &tmp
		case "gif":
			tmp := Gif
			encodeTo = &tmp
		case "bmp":
			tmp := Bmp
			encodeTo = &tmp
		case "tiff":
			tmp := Tiff
			encodeTo = &tmp
		}
	}

	var longestSide uint
	if req.LongestSide != nil {
		longestSide = *req.LongestSide
	} else {
		longestSide = 0
	}

	var obfuscate bool
	if req.Obfuscate != nil {
		obfuscate = *req.Obfuscate
	} else {
		obfuscate = false
	}

	return &ConversionOp{
		Suffix:      req.Suffix,
		EncodeTo:    encodeTo,
		LongestSide: longestSide,
		ResizeOp:    resizeOp,
		Obfuscate:   obfuscate,
	}
}

func makeOriginalOp() *ConversionOp {
	return &ConversionOp{
		Suffix:   "original",
		ResizeOp: Original,
	}
}

// TODO define new imageData structs as a series of operations rather than saving everything
// in memory.

// imageData is a generic image container that accepts raw image data, converts to the go
// image.Image format and uses that to encode new versions. This new format with increased
// metadata will allow for a generic container that can transcode from one format to another.
type imageData struct {
	OriginalImageType ImageType
	OriginalData      []byte
	ImageData         *image.Image
	ExifData          *exifData
	Orientation       Orientation
}

// Checks the EncodeTo parameter. If it's specified, it uses that image format to
// encode the image. If it's not specified, it encodes using the OriginalImageType format.
func (dat *imageData) EncodeImage(op *ConversionOp) ([]byte, *ImageSize, error) {
	if op.ResizeOp == Original && dat.OriginalData != nil && len(dat.OriginalData) > 0 {
		return dat.OriginalData, GetImageSize(dat.ImageData), nil
	}

	var outputImage *image.Image

	if op.ResizeOp == Thumbnail {
		outputImage = dat.MakeThumbnail()
	} else if op.ResizeOp == Scale && op.LongestSide > 0 {
		outputImage = dat.ResizeImage(op.LongestSide)
	} else if op.ResizeOp == ScaleByWidth && op.LongestSide > 0 {
		outputImage = dat.ResizeImageByWidth(op.LongestSide)
	} else {
		outputImage = dat.ImageData
	}

	var encType ImageType
	if op.EncodeTo != nil {
		encType = *op.EncodeTo
	} else {
		encType = dat.OriginalImageType
	}

	switch encType {
	case Jpeg:
		return (*dat).EncodeJpegImage(outputImage)
	case Png:
		return (*dat).EncodePngImage(outputImage)
	case Gif:
		return (*dat).EncodeGifImage(outputImage)
	case Bmp:
		return (*dat).EncodeBmpImage(outputImage)
	case Tiff:
		return (*dat).EncodeTiffImage(outputImage)
	default:
		return nil, nil, errors.New("unsupported image format")
	}
}

// These are the functions that actually perform the encoding operations.
func (dat *imageData) EncodeJpegImage(imgDat *image.Image) ([]byte, *ImageSize, error) {
	var writer io.Writer
	buffer := new(bytes.Buffer)

	// if exif data exists, we'll make an exif writer to encode the jpeg file
	// with the exif data. Otherwise, we'll just use the buffer
	if dat.ExifData != nil {
		writer, _ = newWriterExif(buffer, dat.ExifData)
	} else {
		writer = buffer
	}

	encodeErr := jpeg.Encode(writer, *imgDat, &jpeg.Options{
		Quality: getJpegQuality(),
	})

	if encodeErr != nil {
		return nil, nil, encodeErr
	}

	return buffer.Bytes(), GetImageSize(imgDat), nil
}

func (dat *imageData) EncodePngImage(imgDat *image.Image) ([]byte, *ImageSize, error) {
	enc := png.Encoder{
		CompressionLevel: png.BestCompression,
	}

	buffer := new(bytes.Buffer)

	encodeErr := enc.Encode(buffer, *imgDat)

	if encodeErr != nil {
		return nil, nil, encodeErr
	}

	return buffer.Bytes(), GetImageSize(imgDat), nil
}

func (dat *imageData) EncodeGifImage(imgDat *image.Image) ([]byte, *ImageSize, error) {
	buffer := new(bytes.Buffer)

	encodeErr := gif.Encode(buffer, *imgDat, nil)

	if encodeErr != nil {
		return nil, nil, encodeErr
	}

	return buffer.Bytes(), GetImageSize(imgDat), nil
}

func (dat *imageData) EncodeBmpImage(imgDat *image.Image) ([]byte, *ImageSize, error) {
	buffer := new(bytes.Buffer)

	encodeErr := bmp.Encode(buffer, *imgDat)

	if encodeErr != nil {
		return nil, nil, encodeErr
	}

	return buffer.Bytes(), GetImageSize(imgDat), nil
}

func (dat *imageData) EncodeTiffImage(imgDat *image.Image) ([]byte, *ImageSize, error) {
	buffer := new(bytes.Buffer)

	// encodeErr := tiff.Encode(buffer, *td.ImageData, nil)
	encodeErr := tiff.Encode(buffer, *imgDat, &tiff.Options{
		Compression: tiff.Deflate,
	})

	if encodeErr != nil {
		return nil, nil, encodeErr
	}

	return buffer.Bytes(), GetImageSize(imgDat), nil
}

func (dat *imageData) MakeThumbnail() *image.Image {
	newImage := makeThumbnail(dat.ImageData)

	return newImage
}

// Creates a new imageData struct from the existing struct with a different image size.
// Also allows the user to define a new output format.
func (dat *imageData) ResizeImage(longestSide uint) *image.Image {
	newImage := scaleImage(dat.ImageData, longestSide)

	return newImage
}

// Creates a new imageData struct from the existing struct with a different image size.
// Also allows the user to define a new output format.
func (dat *imageData) ResizeImageByWidth(width uint) *image.Image {
	rotated := dat.Orientation == RotateCCW || dat.Orientation == RotateCW
	newImage := scaleImageByWidth(dat.ImageData, width, rotated)

	return newImage
}

func makeImageDataFromBytes(imageBytes []byte) (*imageData, error) {
	originalImage, t, imageErr := image.Decode(bytes.NewReader(imageBytes))

	if imageErr != nil {
		return nil, imageErr
	}

	var iType ImageType
	var exifDat *exifData
	var orientation Orientation = Horizontal
	switch t {
	case "jpeg":
		iType = Jpeg
		exifDat = extractJpegExif(imageBytes)
		orientation = exifDat.isImageRotated()
	case "png":
		iType = Png
	case "gif":
		iType = Gif
	case "bmp":
		iType = Bmp
	case "tiff":
		iType = Tiff
	default:
		return nil, errors.New("invalid image format")
	}

	return &imageData{
		OriginalImageType: iType,
		OriginalData:      imageBytes,
		ImageData:         &originalImage,
		ExifData:          exifDat,
		Orientation:       orientation,
	}, nil
}

func makeImageDataFromImage(imgDat *image.Image, iType ImageType, exifDat *exifData) *imageData {
	orientation := exifDat.isImageRotated()
	return &imageData{
		OriginalImageType: iType,
		ImageData:         imgDat,
		ExifData:          exifDat,
		Orientation:       orientation,
	}
}

func GetImageSize(imgData *image.Image) *ImageSize {
	bounds := (*imgData).Bounds()

	return &ImageSize{
		Width:  bounds.Max.X,
		Height: bounds.Max.Y,
	}
}
