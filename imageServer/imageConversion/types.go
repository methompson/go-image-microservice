package imageConversion

import (
	"bytes"
	"image"
	gif "image/gif"
	jpeg "image/jpeg"
	png "image/png"
	"io"

	bmp "golang.org/x/image/bmp"
	tiff "golang.org/x/image/tiff"
	"methompson.com/image-microservice/imageServer/dbController"
)

/****************************************************************************************
 * ImageOutputData
*****************************************************************************************/

// The eventual data struct that communicates the result of having written files to the
// filesystem. It provides information, like, name, extension and size formats
type ImageOutputData struct {
	Name        string
	SizeFormats []*dbController.ImageSizeFormat
	Extension   string
}

// Convenience function for exporting a consistent file name.
func (iod *ImageOutputData) MakeFileName(suffix string) string {
	return iod.Name + "@" + suffix + "." + iod.Extension
}

func (iod *ImageOutputData) AddSizeFormat(sf *dbController.ImageSizeFormat) {
	iod.SizeFormats = append(iod.SizeFormats, sf)
}

func makeImageOutputData(bundle *imageBundle) *ImageOutputData {
	return &ImageOutputData{
		Name:        bundle.Name,
		Extension:   bundle.Extension,
		SizeFormats: make([]*dbController.ImageSizeFormat, 0),
	}
}

/****************************************************************************************
 * imageBundle
*****************************************************************************************/

// imageBundle is a bunch of meta information, plus image data for exporting
// a new image file.
type imageBundle struct {
	Name      string
	Extension string
	imageData []*imageData
}

// Convenience function for exporting a consistent file name.
func (iwd *imageBundle) MakeFileName(suffix string) string {
	return iwd.Name + "@" + suffix + "." + iwd.Extension
}

// Creates a new imageWriteDataSize struct and appends it to the writeData slice
func (iwd *imageBundle) addNewSize(dat *imageData) {
	iwd.imageData = append(iwd.imageData, dat)
}

// Helper function for creating a new imageBundle struct
func makeWriteData(name, extension string) *imageBundle {
	return &imageBundle{
		Name:      name,
		Extension: extension,
		imageData: make([]*imageData, 0),
	}
}

// Convenience function for making an imageWriteData struct with jpeg filled in
func makeJpegBundle(name string) *imageBundle {
	return makeWriteData(name, "jpg")
}

// Convenience function for making an imageWriteData struct with png filled in
func makePngBundle(name string) *imageBundle {
	return makeWriteData(name, "png")
}

// Convenience function for making an imageWriteData struct with gif filled in
func makeGifBundle(name string) *imageBundle {
	return makeWriteData(name, "gif")
}

// Convenience function for making an imageWriteData struct with bmp filled in
func makeBmpBundle(name string) *imageBundle {
	return makeWriteData(name, "bmp")
}

// Convenience function for making an imageWriteData struct with tiff filled in
func makeTiffBundle(name string) *imageBundle {
	return makeWriteData(name, "tiff")
}

/****************************************************************************************
 * imageData
*****************************************************************************************/
// Interface for representing image data for conversion. Should be overriden for each
// file format. Each overriden child class needs to contain whatever data is required
// to make EncodeImage with the below signature work.
type imageData interface {
	// EncodeImage converts the image data into the bytes for an image file
	EncodeImage() ([]byte, error)
	GetImageSize() *dbController.ImageSize
	GetSuffix() string
	GetImageData() *image.Image
}

func getBounds(img *image.Image) *dbController.ImageSize {
	bounds := (*img).Bounds()

	return &dbController.ImageSize{
		Width:  bounds.Max.X,
		Height: bounds.Max.Y,
	}
}

/****************************************************************************************
 * jpegData
*****************************************************************************************/
type jpegData struct {
	Suffix       string
	ExifData     *exifData
	ImageData    *image.Image
	OriginalData []byte
}

// When encoding, if original data is included, we'll just return that with no further
// processing or encoding. If originalData is not included, we encode the ImageData
func (dat *jpegData) EncodeImage() ([]byte, error) {
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

func (dat *jpegData) GetImageSize() *dbController.ImageSize {
	return getBounds(dat.ImageData)
}

func (dat *jpegData) GetSuffix() string {
	return dat.Suffix
}

func (dat *jpegData) GetImageData() *image.Image {
	return dat.ImageData
}

func (dat *jpegData) AsImageData() *imageData {
	var imgData imageData = dat
	return &imgData
}

// Convenience function that takes jpeg file bytes, extract exif data and decodes the
// image file. This function also returns the original image data. We don't worry about
// performing any image resizes yet.
func makeJpegData(imageBytes []byte, suffix string) (*jpegData, error) {
	exifData := extractJpegExif(imageBytes)

	originalImage, _, imageErr := image.Decode(bytes.NewReader(imageBytes))

	if imageErr != nil {
		return nil, imageErr
	}

	return &jpegData{
		Suffix:       suffix,
		ImageData:    &originalImage,
		ExifData:     exifData,
		OriginalData: imageBytes,
	}, nil
}

// Replaces the imageData with a resized version of the image data. Makes a new jpegData
// struct, discards the original bytes
func makeScaledJpegData(iData *jpegData, longestSide uint, suffix string) *jpegData {
	newImage := scaleImage(iData.ImageData, longestSide)

	return &jpegData{
		Suffix:    suffix,
		ImageData: newImage,
		ExifData:  iData.ExifData,
	}
}

// Replaces the imageData with a resized version of the image data. Makes a new jpegData
// struct, discards the original bytes
func makeThumbnailJpegData(iData *jpegData, suffix string) *jpegData {
	newImage := makeThumbnail(iData.ImageData)

	return &jpegData{
		Suffix:    suffix,
		ImageData: newImage,
		ExifData:  iData.ExifData,
	}
}

/****************************************************************************************
 * pngData
*****************************************************************************************/
type pngData struct {
	Suffix       string
	ImageData    *image.Image
	OriginalData []byte
}

func (dat *pngData) EncodeImage() ([]byte, error) {
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

func (dat *pngData) GetImageSize() *dbController.ImageSize {
	return getBounds(dat.ImageData)
}

func (dat *pngData) GetSuffix() string {
	return dat.Suffix
}

func (dat *pngData) GetImageData() *image.Image {
	return dat.ImageData
}

func (dat *pngData) AsImageData() *imageData {
	var imgData imageData = dat
	return &imgData
}

func makePngData(imageBytes []byte, suffix string) (*pngData, error) {
	originalImage, _, imageErr := image.Decode(bytes.NewReader(imageBytes))

	if imageErr != nil {
		return nil, imageErr
	}

	return &pngData{
		Suffix:       suffix,
		ImageData:    &originalImage,
		OriginalData: imageBytes,
	}, nil
}

func makeScaledPngData(iData *pngData, longestSide uint, suffix string) *pngData {
	newImage := scaleImage(iData.ImageData, longestSide)

	return &pngData{
		Suffix:    suffix,
		ImageData: newImage,
	}
}

func makeThumbnailPngData(iData *pngData, suffix string) *pngData {
	newImage := makeThumbnail(iData.ImageData)

	return &pngData{
		Suffix:    suffix,
		ImageData: newImage,
	}
}

/****************************************************************************************
 * gifData
*****************************************************************************************/
type gifData struct {
	Suffix       string
	ImageData    *image.Image
	OriginalData []byte
}

func (dat *gifData) EncodeImage() ([]byte, error) {
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

func (dat *gifData) GetImageSize() *dbController.ImageSize {
	return getBounds(dat.ImageData)
}

func (dat *gifData) GetSuffix() string {
	return dat.Suffix
}

func (dat *gifData) GetImageData() *image.Image {
	return dat.ImageData
}

func (dat *gifData) AsImageData() *imageData {
	var imgData imageData = dat
	return &imgData
}

func makeGifData(imageBytes []byte, suffix string) (*gifData, error) {
	originalImage, _, imageErr := image.Decode(bytes.NewReader(imageBytes))

	if imageErr != nil {
		return nil, imageErr
	}

	return &gifData{
		Suffix:       suffix,
		ImageData:    &originalImage,
		OriginalData: imageBytes,
	}, nil
}

func makeScaledGifData(iData *gifData, longestSide uint, suffix string) *gifData {
	newImage := scaleImage(iData.ImageData, longestSide)

	return &gifData{
		Suffix:    suffix,
		ImageData: newImage,
	}
}

func makeThumbnailGifData(iData *gifData, suffix string) *gifData {
	newImage := makeThumbnail(iData.ImageData)

	return &gifData{
		Suffix:    suffix,
		ImageData: newImage,
	}
}

/****************************************************************************************
 * bmpData
*****************************************************************************************/
type bmpData struct {
	Suffix       string
	ImageData    *image.Image
	OriginalData []byte
}

func (dat *bmpData) EncodeImage() ([]byte, error) {
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

func (dat *bmpData) GetImageSize() *dbController.ImageSize {
	return getBounds(dat.ImageData)
}

func (dat *bmpData) GetSuffix() string {
	return dat.Suffix
}

func (dat *bmpData) GetImageData() *image.Image {
	return dat.ImageData
}

func (dat *bmpData) AsImageData() *imageData {
	var imgData imageData = dat
	return &imgData
}

func makeBmpData(imageBytes []byte, suffix string) (*bmpData, error) {
	originalImage, _, imageErr := image.Decode(bytes.NewReader(imageBytes))

	if imageErr != nil {
		return nil, imageErr
	}

	return &bmpData{
		Suffix:       suffix,
		ImageData:    &originalImage,
		OriginalData: imageBytes,
	}, nil
}

func makeScaledBmpData(iData *bmpData, longestSide uint, suffix string) *bmpData {
	newImage := scaleImage(iData.ImageData, longestSide)

	return &bmpData{
		Suffix:    suffix,
		ImageData: newImage,
	}
}

func makeThumbnailBmpData(iData *bmpData, suffix string) *bmpData {
	newImage := makeThumbnail(iData.ImageData)

	return &bmpData{
		Suffix:    suffix,
		ImageData: newImage,
	}
}

/****************************************************************************************
 * tiffData
*****************************************************************************************/
type tiffData struct {
	Suffix       string
	ImageData    *image.Image
	OriginalData []byte
}

func (dat *tiffData) EncodeImage() ([]byte, error) {
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

func (dat *tiffData) GetImageSize() *dbController.ImageSize {
	return getBounds(dat.ImageData)
}

func (dat *tiffData) GetSuffix() string {
	return dat.Suffix
}

func (dat *tiffData) GetImageData() *image.Image {
	return dat.ImageData
}

func (dat *tiffData) AsImageData() *imageData {
	var imgData imageData = dat
	return &imgData
}

func makeTiffData(imageBytes []byte, suffix string) (*tiffData, error) {
	originalImage, _, imageErr := image.Decode(bytes.NewReader(imageBytes))

	if imageErr != nil {
		return nil, imageErr
	}

	return &tiffData{
		Suffix:       suffix,
		ImageData:    &originalImage,
		OriginalData: imageBytes,
	}, nil
}

func makeScaledTiffData(iData *tiffData, longestSide uint, suffix string) *tiffData {
	newImage := scaleImage(iData.ImageData, longestSide)

	return &tiffData{
		Suffix:    suffix,
		ImageData: newImage,
	}
}

func makeThumbnailTiffData(iData *tiffData, suffix string) *tiffData {
	newImage := makeThumbnail(iData.ImageData)

	return &tiffData{
		Suffix:    suffix,
		ImageData: newImage,
	}
}
