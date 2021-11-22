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

func ProcessImageFile(ctx *gin.Context) (*ImageOutputData, error) {
	file, fileHandler, fileErr := ctx.Request.FormFile("image")

	if fileErr != nil {
		return nil, fileErr
	}
	defer file.Close()

	contentType := fileHandler.Header.Get("Content-Type")

	fileBytes, fileBytesErr := ioutil.ReadAll(file)

	if fileBytesErr != nil {
		return nil, fileBytesErr
	}

	switch contentType {
	case "image/heic":
		return processNewHeifImage(fileBytes)
	case "image/jpeg":
		return processNewJpegImage(fileBytes)
	case "image/png":
		return processNewPngImage(fileBytes)
	case "image/gif":
		return processNewGifImage(fileBytes)
	case "image/bmp":
		return processNewBmpImage(fileBytes)
	case "image/tiff":
		return processNewTiffImage(fileBytes)
	default:
		return nil, errors.New("invalid image format")
	}
}

// Returns a string to be used as a file name. Currently just uses UUID
func makeName() string {
	return uuid.New().String()
}

// Takes a slice of imageWriteData objects and sends each one to the writeFileData
// function. If a write error occurs, the process is halted and an error is thrown
func writeFiles(writeData []*imageWriteData) error {
	var err error
	for _, w := range writeData {
		writeErr := writeFileData(w)
		if writeErr != nil {
			err = writeErr
		}
	}

	if err != nil {
		rollBackWrites(writeData)
		return err
	}

	return nil
}

// writeFileData takes the imgWriteData object, gets an image path, creates the
// image path, if it doesn't exist, encodes the image and then writes the resultant
// bytes to the file system. Write errors are returned
func writeFileData(imgData *imageWriteData) error {
	folderPath := GetImagePath(imgData.Name)

	folderErr := CheckOrCreateImageFolder(folderPath)

	if folderErr != nil {
		return folderErr
	}

	filePath := path.Join(folderPath, imgData.MakeFileName())
	bytes, encodeErr := (*imgData.ImageData).EncodeImage()

	if encodeErr != nil {
		return encodeErr
	}

	return os.WriteFile(filePath, bytes, 0644)
}

// Attempts to roll back any writes that already occurred in the case of an error
func rollBackWrites(writeData []*imageWriteData) {
	for _, w := range writeData {
		folderPath := GetImagePath(w.Name)
		filePath := path.Join(folderPath, w.MakeFileName())

		deleteFile(filePath)
	}
}

// Simple file deletion function.
func deleteFile(filePath string) error {
	return os.Remove(filePath)
}

// The commonFileProcess generates the ImageOutputData object from the elements
// generated in each process function. This will be the data that we eventually
// save to our database.
func commonFileProcess(writeData []*imageWriteData, filename, extension string) *ImageOutputData {
	suffixes := make([]string, 0)
	for _, el := range writeData {
		suffixes = append(suffixes, el.Suffix)
	}

	return &ImageOutputData{
		Name:      filename,
		Suffixes:  suffixes,
		Extension: extension,
	}
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
func processNewHeifImage(imageBytes []byte) (*ImageOutputData, error) {
	reader := bytes.NewReader(imageBytes)
	exif, err := goheif.ExtractExif(reader)
	if err != nil {
		return nil, err
	}

	image, err := goheif.Decode(reader)
	if err != nil {
		return nil, err
	}

	jpegData := jpegData{
		ImageData: &image,
		ExifData: &exifData{
			ExifData: exif,
		},
	}

	newBytes, err := jpegData.EncodeImage()

	if err != nil {
		return nil, err
	}

	return processNewJpegImage(newBytes)
}

func processNewJpegImage(imageBytes []byte) (*ImageOutputData, error) {
	writeData := make([]*imageWriteData, 0)
	filename := makeName()

	imageData, imageErr := makeJpegData(imageBytes)

	if imageErr != nil {
		return nil, imageErr
	}

	writeData = append(writeData, makeJpegWriteData(imageData, filename, "original"))
	writeData = append(writeData, makeJpegWriteData(makeThumbnailJpegData(imageData), filename, "thumb"))

	outputData := commonFileProcess(writeData, filename, "jpg")

	writeErr := writeFiles(writeData)

	if writeErr != nil {
		return nil, writeErr
	}

	return outputData, nil
}

func processNewPngImage(imageBytes []byte) (*ImageOutputData, error) {
	writeData := make([]*imageWriteData, 0)
	filename := makeName()

	imageData, imageErr := makePngData(imageBytes)

	if imageErr != nil {
		return nil, imageErr
	}

	writeData = append(writeData, makePngWriteData(imageData, filename, "original"))
	writeData = append(writeData, makePngWriteData(makeThumbnailPngData(imageData), filename, "thumb"))

	outputData := commonFileProcess(writeData, filename, "png")

	writeErr := writeFiles(writeData)

	if writeErr != nil {
		return nil, writeErr
	}

	return outputData, nil
}

// TODO provide the ability to convert from bmp to png or jpg
func processNewGifImage(imageBytes []byte) (*ImageOutputData, error) {
	writeData := make([]*imageWriteData, 0)
	filename := makeName()

	imageData, imageErr := makeGifData(imageBytes)

	if imageErr != nil {
		return nil, imageErr
	}

	writeData = append(writeData, makeGifWriteData(imageData, filename, "original"))
	writeData = append(writeData, makeGifWriteData(makeThumbnailGifData(imageData), filename, "thumb"))

	outputData := commonFileProcess(writeData, filename, "gif")

	writeErr := writeFiles(writeData)

	if writeErr != nil {
		return nil, writeErr
	}

	return outputData, nil
}

// TODO provide the ability to convert from bmp to png or jpg
func processNewBmpImage(imageBytes []byte) (*ImageOutputData, error) {
	writeData := make([]*imageWriteData, 0)
	filename := makeName()

	imageData, imageErr := makeBmpData(imageBytes)

	if imageErr != nil {
		return nil, imageErr
	}

	writeData = append(writeData, makeBmpWriteData(imageData, filename, "original"))
	writeData = append(writeData, makeBmpWriteData(makeThumbnailBmpData(imageData), filename, "thumb"))

	outputData := commonFileProcess(writeData, filename, "bmp")

	writeErr := writeFiles(writeData)

	if writeErr != nil {
		return nil, writeErr
	}

	return outputData, nil
}

// TODO compress tiff images to jpeg
func processNewTiffImage(imageBytes []byte) (*ImageOutputData, error) {
	writeData := make([]*imageWriteData, 0)
	filename := makeName()

	imageData, imageErr := makeTiffData(imageBytes)

	if imageErr != nil {
		return nil, imageErr
	}

	writeData = append(writeData, makeTiffWriteData(imageData, filename, "original"))
	writeData = append(writeData, makeTiffWriteData(makeThumbnailTiffData(imageData), filename, "thumb"))

	outputData := commonFileProcess(writeData, filename, "tiff")

	writeErr := writeFiles(writeData)

	if writeErr != nil {
		return nil, writeErr
	}

	return outputData, nil
}
