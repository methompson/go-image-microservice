package imageConversion

import (
	"bytes"
	"errors"
	"image"
	"io"

	gif "image/gif"
	jpeg "image/jpeg"
	png "image/png"

	bmp "golang.org/x/image/bmp"
	tiff "golang.org/x/image/tiff"
)

type ImageType int64

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

// imageData is a generic image container that accepts raw image data, converts to the go
// image.Image format and uses that to encode new versions. This new format with increased
// metadata will allow for a generic container that can transcode from one format to another.
type imageData struct {
	OriginalImageType ImageType
	EncodeTo          ImageType
	OriginalData      []byte
	ImageData         *image.Image
	ExifData          *exifData
	Suffix            string
	Orientation       Orientation
}

func (dat *imageData) GetExtension() string {
	switch dat.EncodeTo {
	case Jpeg:
		return "jpg"
	case Png:
		return "png"
	case Gif:
		return "gif"
	case Bmp:
		return "bmp"
	case Tiff:
		return "tiff"
	default:
		return ""
	}
}

// Checks the EncodeTo parameter. If it's specified, it uses that image format to
// encode the image. If it's not specified, it encodes using the OriginalImageType format.
func (dat *imageData) EncodeImage() ([]byte, error) {
	switch dat.EncodeTo {
	case Jpeg:
		return (*dat).EncodeJpegImage()
	case Png:
		return (*dat).EncodePngImage()
	case Gif:
		return (*dat).EncodeGifImage()
	case Bmp:
		return (*dat).EncodeBmpImage()
	case Tiff:
		return (*dat).EncodeTiffImage()
	default:
		return nil, errors.New("unsupported image format")
	}
}

// These are the functions that actually perform the encoding operations.
func (dat *imageData) EncodeJpegImage() ([]byte, error) {
	if dat.OriginalData != nil {
		return dat.OriginalData, nil
	}

	var writer io.Writer
	buffer := new(bytes.Buffer)

	// if exif data exists, we'll make an exif writer to encode the jpeg file
	// with the exif data. Otherwise, we'll just use the buffer
	if dat.ExifData != nil {
		writer, _ = newWriterExif(buffer, dat.ExifData)
	} else {
		writer = buffer
	}

	encodeErr := jpeg.Encode(writer, *dat.ImageData, &jpeg.Options{
		Quality: getJpegQuality(),
	})

	if encodeErr != nil {
		return nil, encodeErr
	}

	return buffer.Bytes(), nil
}

func (dat *imageData) EncodePngImage() ([]byte, error) {
	if dat.OriginalData != nil {
		return dat.OriginalData, nil
	}

	enc := png.Encoder{
		CompressionLevel: png.BestCompression,
	}

	buffer := new(bytes.Buffer)

	encodeErr := enc.Encode(buffer, *dat.ImageData)

	if encodeErr != nil {
		return nil, encodeErr
	}

	return buffer.Bytes(), nil
}

func (dat *imageData) EncodeGifImage() ([]byte, error) {
	if dat.OriginalData != nil {
		return dat.OriginalData, nil
	}

	buffer := new(bytes.Buffer)

	encodeErr := gif.Encode(buffer, *dat.ImageData, nil)

	if encodeErr != nil {
		return nil, encodeErr
	}

	return buffer.Bytes(), nil
}

func (dat *imageData) EncodeBmpImage() ([]byte, error) {
	if dat.OriginalData != nil {
		return dat.OriginalData, nil
	}

	buffer := new(bytes.Buffer)

	encodeErr := bmp.Encode(buffer, *dat.ImageData)

	if encodeErr != nil {
		return nil, encodeErr
	}

	return buffer.Bytes(), nil
}

func (dat *imageData) EncodeTiffImage() ([]byte, error) {
	if dat.OriginalData != nil {
		return dat.OriginalData, nil
	}

	buffer := new(bytes.Buffer)

	// encodeErr := tiff.Encode(buffer, *td.ImageData, nil)
	encodeErr := tiff.Encode(buffer, *dat.ImageData, &tiff.Options{
		Compression: tiff.Deflate,
	})

	if encodeErr != nil {
		return nil, encodeErr
	}

	return buffer.Bytes(), nil
}

func (dat *imageData) EncodeToJpeg() { dat.EncodeTo = Jpeg }
func (dat *imageData) EncodeToPng()  { dat.EncodeTo = Png }
func (dat *imageData) EncodeToGif()  { dat.EncodeTo = Gif }
func (dat *imageData) EncodeToBmp()  { dat.EncodeTo = Bmp }
func (dat *imageData) EncodeToTiff() { dat.EncodeTo = Tiff }

func (dat *imageData) MakeThumbnail() *imageData {
	newImage := makeThumbnail(dat.ImageData)

	return &imageData{
		OriginalImageType: dat.OriginalImageType,
		EncodeTo:          dat.EncodeTo,
		ImageData:         newImage,
		ExifData:          dat.ExifData,
		Suffix:            "thumb",
		Orientation:       dat.Orientation,
	}
}

// Creates a new imageData struct from the existing struct with a different image size.
// Also allows the user to define a new output format.
func (dat *imageData) ResizeImage(suffix string, longestSide uint) *imageData {
	newImage := scaleImage(dat.ImageData, longestSide)

	return &imageData{
		OriginalImageType: dat.OriginalImageType,
		EncodeTo:          dat.EncodeTo,
		ImageData:         newImage,
		ExifData:          dat.ExifData,
		Suffix:            suffix,
		Orientation:       dat.Orientation,
	}
}

// Creates a new imageData struct from the existing struct with a different image size.
// Also allows the user to define a new output format.
func (dat *imageData) ResizeImageByWidth(suffix string, width uint) *imageData {
	rotated := dat.Orientation == RotateCCW || dat.Orientation == RotateCW
	newImage := scaleImageByWidth(dat.ImageData, width, rotated)

	return &imageData{
		OriginalImageType: dat.OriginalImageType,
		EncodeTo:          dat.EncodeTo,
		ImageData:         newImage,
		ExifData:          dat.ExifData,
		Suffix:            suffix,
		Orientation:       dat.Orientation,
	}
}

func (dat *imageData) GetImageSize() *ImageSize {
	bounds := (*dat.ImageData).Bounds()

	return &ImageSize{
		Width:  bounds.Max.X,
		Height: bounds.Max.Y,
	}
}

func makeImageDataFromBytes(imageBytes []byte, suffix string) (*imageData, error) {
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
		EncodeTo:          iType,
		OriginalData:      imageBytes,
		ImageData:         &originalImage,
		ExifData:          exifDat,
		Suffix:            suffix,
		Orientation:       orientation,
	}, nil
}

func makeImageDataFromImage(imgDat *image.Image, iType ImageType, suffix string, exifDat *exifData) *imageData {
	orientation := exifDat.isImageRotated()
	return &imageData{
		OriginalImageType: iType,
		EncodeTo:          iType,
		ImageData:         imgDat,
		ExifData:          exifDat,
		Suffix:            suffix,
		Orientation:       orientation,
	}
}
