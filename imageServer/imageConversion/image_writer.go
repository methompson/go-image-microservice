package imageConversion

import (
	"bytes"
	"errors"
	_ "image/gif"
	"io/ioutil"

	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jdeng/goheif"
)

func SaveImageFile(ctx *gin.Context) error {
	file, fileHandler, fileErr := ctx.Request.FormFile("image")

	if fileErr != nil {
		return fileErr
	}
	defer file.Close()

	contentType := fileHandler.Header.Get("Content-Type")

	fileBytes, fileBytesErr := ioutil.ReadAll(file)

	if fileBytesErr != nil {
		return fileBytesErr
	}

	switch contentType {
	case "image/heic":
		return processNewHeifImage(fileBytes)
	case "image/jpeg":
		return processNewJpegImage(fileBytes)
	case "image/png":
		return processNewPngImage(fileBytes)
	case "image/gif":
		return processGifImage(fileBytes)
	case "image/bmp":
		return processBmpImage(fileBytes)
	case "image/tiff":
		return processTiffImage(fileBytes)
	default:
		return errors.New("invalid image format")
	}
}

func makeName() string {
	return uuid.New().String()
}

func writeFileData(folderPath string, imgData *imageWriteData) error {
	filePath := path.Join(folderPath, imgData.MakeFileName())
	bytes, encodeErr := (*imgData.ImageData).EncodeImage()

	if encodeErr != nil {
		return encodeErr
	}

	writeErr := os.WriteFile(filePath, bytes, 0644)

	if writeErr != nil {
		return writeErr
	}

	return nil
}

// The save functions need to do a few things:
// * They need to save the original file to the server
// * They need to perform whatever resize conversions that are prescribed by the environment
// * They need to perform a filetype conversion if it's not easliy renderable by browsers

// For processing heif images for web use, we need to convert to a web-readable image.
// By default, we will convert the image to jpeg. The process will involve the following:
// * Get an *image.Image object
// * Get the exif
// Then we pass the above two points to the encode Jpeg function.
func processNewHeifImage(imageBytes []byte) error {
	reader := bytes.NewReader(imageBytes)
	exif, err := goheif.ExtractExif(reader)
	if err != nil {
		return err
	}

	image, err := goheif.Decode(reader)
	if err != nil {
		return err
	}

	jpegData := jpegData{
		ImageData: &image,
		ExifData: &exifData{
			ExifData: exif,
		},
	}

	newBytes, err := jpegData.EncodeImage()

	if err != nil {
		return err
	}

	return processNewJpegImage(newBytes)
}

func processNewJpegImage(imageBytes []byte) error {
	writeData := make([]*imageWriteData, 0)
	filename := makeName()

	imageData, imageErr := makeJpegData(imageBytes)

	if imageErr != nil {
		return imageErr
	}

	writeData = append(writeData, makeJpegWriteData(imageData, filename, "original"))
	writeData = append(writeData, makeJpegWriteData(makeThumbnailJpegData(imageData), filename, "thumb"))

	for _, w := range writeData {
		writeErr := writeFileData("./files", w)
		if writeErr != nil {
			return writeErr
		}
	}

	return nil
}

func processNewPngImage(imageBytes []byte) error {
	writeData := make([]*imageWriteData, 0)
	filename := makeName()

	imageData, imageErr := makePngData(imageBytes)

	if imageErr != nil {
		return imageErr
	}

	writeData = append(writeData, makePngWriteData(imageData, filename, "original"))
	writeData = append(writeData, makePngWriteData(makeThumbnailPngData(imageData), filename, "thumb"))

	for _, w := range writeData {
		writeErr := writeFileData("./files", w)
		if writeErr != nil {
			return writeErr
		}
	}

	return nil
}

// TODO provide the ability to convert from bmp to png or jpg
func processGifImage(imageBytes []byte) error {
	writeData := make([]*imageWriteData, 0)
	filename := makeName()

	imageData, imageErr := makeGifData(imageBytes)

	if imageErr != nil {
		return imageErr
	}

	writeData = append(writeData, makeGifWriteData(imageData, filename, "original"))
	writeData = append(writeData, makeGifWriteData(makeThumbnailGifData(imageData), filename, "thumb"))

	for _, w := range writeData {
		writeErr := writeFileData("./files", w)
		if writeErr != nil {
			return writeErr
		}
	}

	return nil
}

// TODO provide the ability to convert from bmp to png or jpg
func processBmpImage(imageBytes []byte) error {
	writeData := make([]*imageWriteData, 0)
	filename := makeName()

	imageData, imageErr := makeBmpData(imageBytes)

	if imageErr != nil {
		return imageErr
	}

	writeData = append(writeData, makeBmpWriteData(imageData, filename, "original"))
	writeData = append(writeData, makeBmpWriteData(makeThumbnailBmpData(imageData), filename, "thumb"))

	for _, w := range writeData {
		writeErr := writeFileData("./files", w)
		if writeErr != nil {
			return writeErr
		}
	}

	return nil
}

// TODO compress tiff images to jpeg
func processTiffImage(imageBytes []byte) error {
	writeData := make([]*imageWriteData, 0)
	filename := makeName()

	imageData, imageErr := makeTiffData(imageBytes)

	if imageErr != nil {
		return imageErr
	}

	writeData = append(writeData, makeTiffWriteData(imageData, filename, "original"))
	writeData = append(writeData, makeTiffWriteData(makeThumbnailTiffData(imageData), filename, "thumb"))

	for _, w := range writeData {
		writeErr := writeFileData("./files", w)
		if writeErr != nil {
			return writeErr
		}
	}

	return nil
}
