package imageConversion

import (
	"bytes"
	"image"
	gif "image/gif"
	jpeg "image/jpeg"
	png "image/png"

	bmp "golang.org/x/image/bmp"
	tiff "golang.org/x/image/tiff"

	"io"
)

/****************************************************************************************
 * imageWriteData
*****************************************************************************************/
type imageWriteData struct {
	Name      string
	Suffix    string
	extension string
	ImageData *imageData
}

func (iwd *imageWriteData) MakeFileName() string {
	return iwd.Name + "@" + iwd.Suffix + "." + iwd.extension
}

func makeJpegWriteData(jpgData *jpegData, name string, suffix string) *imageWriteData {
	var imgData imageData = jpgData

	return &imageWriteData{
		Name:      name,
		Suffix:    suffix,
		extension: "jpg",
		ImageData: &imgData,
	}
}

func makePngWriteData(pData *pngData, name string, suffix string) *imageWriteData {
	var imgData imageData = pData

	return &imageWriteData{
		Name:      name,
		Suffix:    suffix,
		extension: "png",
		ImageData: &imgData,
	}
}

func makeGifWriteData(gData *gifData, name string, suffix string) *imageWriteData {
	var imgData imageData = gData

	return &imageWriteData{
		Name:      name,
		Suffix:    suffix,
		extension: "gif",
		ImageData: &imgData,
	}
}

func makeBmpWriteData(bData *bmpData, name string, suffix string) *imageWriteData {
	var imgData imageData = bData

	return &imageWriteData{
		Name:      name,
		Suffix:    suffix,
		extension: "bmp",
		ImageData: &imgData,
	}
}

func makeTiffWriteData(tData *tiffData, name string, suffix string) *imageWriteData {
	var imgData imageData = tData

	return &imageWriteData{
		Name:      name,
		Suffix:    suffix,
		extension: "tiff",
		ImageData: &imgData,
	}
}

/****************************************************************************************
 * exifData
*****************************************************************************************/
type exifData struct {
	ExifData []byte
}

func (exif *exifData) makeSizeData() []byte {
	markerlen := 2 + len(exif.ExifData)

	// The size of the marker is represented as 2 bytes, so we have to convert the
	// length into two bytes to place into the exif marker
	sizeByte1 := uint8(markerlen >> 8)
	sizeByte2 := uint8(markerlen & 0xff)

	exifMarker := []byte{0xff, 0xe1, sizeByte1, sizeByte2}

	return exifMarker
}

func (exif *exifData) makeFileData() []byte {
	data := exif.makeSizeData()
	data = append(data, exif.ExifData...)

	return data
}

/****************************************************************************************
 * imageData
*****************************************************************************************/
type imageData interface {
	// EncodeImage converts the image data into the bytes for an image file
	EncodeImage() ([]byte, error)
}

/****************************************************************************************
 * jpegData
*****************************************************************************************/
type jpegData struct {
	ExifData     *exifData
	ImageData    *image.Image
	OriginalData []byte
}

func (jd *jpegData) EncodeImage() ([]byte, error) {
	if jd.OriginalData != nil {
		return jd.OriginalData, nil
	}

	var writer io.Writer
	buffer := new(bytes.Buffer)

	// if exif data exists, we'll make an exif writer to encode the jpeg file
	// with the exif data. Otherwise, we'll just use the buffer
	if jd.ExifData != nil {
		writer, _ = newWriterExif(buffer, jd.ExifData)
	} else {
		writer = buffer
	}

	encodeErr := jpeg.Encode(writer, *jd.ImageData, &jpeg.Options{
		Quality: 75,
	})

	if encodeErr != nil {
		return nil, encodeErr
	}

	return buffer.Bytes(), nil
}

func makeJpegData(imageBytes []byte) (*jpegData, error) {
	exifData := extractJpegExif(imageBytes)

	originalImage, _, imageErr := image.Decode(bytes.NewReader(imageBytes))

	if imageErr != nil {
		return nil, imageErr
	}

	return &jpegData{
		ImageData:    &originalImage,
		ExifData:     exifData,
		OriginalData: imageBytes,
	}, nil
}

func makeScaledJpegData(jData *jpegData, longestSide uint) *jpegData {
	newImage := scaleImage(jData.ImageData, longestSide)

	return &jpegData{
		ImageData: newImage,
		ExifData:  jData.ExifData,
	}
}

func makeThumbnailJpegData(jData *jpegData) *jpegData {
	newImage := makeThumbnail(jData.ImageData)

	return &jpegData{
		ImageData: newImage,
		ExifData:  jData.ExifData,
	}
}

/****************************************************************************************
 * pngData
*****************************************************************************************/
type pngData struct {
	ImageData    *image.Image
	OriginalData []byte
}

func (pd *pngData) EncodeImage() ([]byte, error) {
	if pd.OriginalData != nil {
		return pd.OriginalData, nil
	}

	enc := png.Encoder{
		CompressionLevel: png.BestCompression,
	}

	buffer := new(bytes.Buffer)

	encodeErr := enc.Encode(buffer, *pd.ImageData)

	if encodeErr != nil {
		return nil, encodeErr
	}

	return buffer.Bytes(), nil
}

func makePngData(imageBytes []byte) (*pngData, error) {
	originalImage, _, imageErr := image.Decode(bytes.NewReader(imageBytes))

	if imageErr != nil {
		return nil, imageErr
	}

	return &pngData{
		ImageData:    &originalImage,
		OriginalData: imageBytes,
	}, nil
}

func makeScaledPngData(pData *pngData, longestSide uint) *pngData {
	newImage := scaleImage(pData.ImageData, longestSide)

	return &pngData{
		ImageData: newImage,
	}
}

func makeThumbnailPngData(pData *pngData) *pngData {
	newImage := makeThumbnail(pData.ImageData)

	return &pngData{
		ImageData: newImage,
	}
}

/****************************************************************************************
 * gifData
*****************************************************************************************/
type gifData struct {
	ImageData    *image.Image
	OriginalData []byte
}

func (gd *gifData) EncodeImage() ([]byte, error) {
	if gd.OriginalData != nil {
		return gd.OriginalData, nil
	}

	buffer := new(bytes.Buffer)

	encodeErr := gif.Encode(buffer, *gd.ImageData, nil)

	if encodeErr != nil {
		return nil, encodeErr
	}

	return buffer.Bytes(), nil
}

func makeGifData(imageBytes []byte) (*gifData, error) {
	originalImage, _, imageErr := image.Decode(bytes.NewReader(imageBytes))

	if imageErr != nil {
		return nil, imageErr
	}

	return &gifData{
		ImageData:    &originalImage,
		OriginalData: imageBytes,
	}, nil
}

func makeScaledGifData(gData *gifData, longestSide uint) *gifData {
	newImage := scaleImage(gData.ImageData, longestSide)

	return &gifData{
		ImageData: newImage,
	}
}

func makeThumbnailGifData(gData *gifData) *gifData {
	newImage := makeThumbnail(gData.ImageData)

	return &gifData{
		ImageData: newImage,
	}
}

/****************************************************************************************
 * bmpData
*****************************************************************************************/
type bmpData struct {
	ImageData    *image.Image
	OriginalData []byte
}

func (bd *bmpData) EncodeImage() ([]byte, error) {
	if bd.OriginalData != nil {
		return bd.OriginalData, nil
	}

	buffer := new(bytes.Buffer)

	encodeErr := bmp.Encode(buffer, *bd.ImageData)

	if encodeErr != nil {
		return nil, encodeErr
	}

	return buffer.Bytes(), nil
}

func makeBmpData(imageBytes []byte) (*bmpData, error) {
	originalImage, _, imageErr := image.Decode(bytes.NewReader(imageBytes))

	if imageErr != nil {
		return nil, imageErr
	}

	return &bmpData{
		ImageData:    &originalImage,
		OriginalData: imageBytes,
	}, nil
}

func makeScaledBmpData(iData *bmpData, longestSide uint) *bmpData {
	newImage := scaleImage(iData.ImageData, longestSide)

	return &bmpData{
		ImageData: newImage,
	}
}

func makeThumbnailBmpData(iData *bmpData) *bmpData {
	newImage := makeThumbnail(iData.ImageData)

	return &bmpData{
		ImageData: newImage,
	}
}

/****************************************************************************************
 * tiffData
*****************************************************************************************/
type tiffData struct {
	ImageData    *image.Image
	OriginalData []byte
}

func (td *tiffData) EncodeImage() ([]byte, error) {
	if td.OriginalData != nil {
		return td.OriginalData, nil
	}

	buffer := new(bytes.Buffer)

	// encodeErr := tiff.Encode(buffer, *td.ImageData, nil)
	encodeErr := tiff.Encode(buffer, *td.ImageData, &tiff.Options{
		Compression: tiff.Deflate,
	})

	if encodeErr != nil {
		return nil, encodeErr
	}

	return buffer.Bytes(), nil
}

func makeTiffData(imageBytes []byte) (*tiffData, error) {
	originalImage, _, imageErr := image.Decode(bytes.NewReader(imageBytes))

	if imageErr != nil {
		return nil, imageErr
	}

	return &tiffData{
		ImageData:    &originalImage,
		OriginalData: imageBytes,
	}, nil
}

func makeScaledTiffData(iData *tiffData, longestSide uint) *tiffData {
	newImage := scaleImage(iData.ImageData, longestSide)

	return &tiffData{
		ImageData: newImage,
	}
}

func makeThumbnailTiffData(iData *tiffData) *tiffData {
	newImage := makeThumbnail(iData.ImageData)

	return &tiffData{
		ImageData: newImage,
	}
}
