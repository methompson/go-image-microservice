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

	"methompson.com/image-microservice/imageServer/dbController"
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
// function. If a write error occurs, the process is halted and an error is returned
func writeFiles(bundle *imageBundle) (*ImageOutputData, error) {
	export := makeImageOutputData(bundle)
	var err error
	for _, imgDat := range bundle.imageData {
		size, filename, writeErr := writeFileData(imgDat, bundle)
		if writeErr != nil {
			err = writeErr
			break
		}

		format := &dbController.ImageSizeFormat{
			Filename:  filename,
			FileSize:  size,
			ImageSize: (*imgDat).GetImageSize(),
		}

		export.AddSizeFormat(format)
	}

	if err != nil {
		undoWrites(bundle)
		return nil, err
	}

	return export, nil
}

// writeFileData takes the imgWriteData object, gets an image path, creates the
// image path, if it doesn't exist, encodes the image and then writes the resultant
// bytes to the file system. Write errors are returned
func writeFileData(writeData *imageData, bundle *imageBundle) (int, string, error) {
	folderPath := GetImagePath(bundle.Name)

	folderErr := CheckOrCreateImageFolder(folderPath)

	if folderErr != nil {
		return -1, "", folderErr
	}

	filename := (*bundle).MakeFileName((*writeData).GetSuffix())

	filePath := path.Join(folderPath, filename)
	bytes, encodeErr := (*writeData).EncodeImage()

	if encodeErr != nil {
		return -1, "", encodeErr
	}

	writeErr := os.WriteFile(filePath, bytes, 0644)

	if writeErr != nil {
		return -1, "", writeErr
	}

	return len(bytes), filename, nil
}

// Attempts to roll back any writes that already occurred in the case of an error
func undoWrites(bundle *imageBundle) error {
	for _, w := range bundle.imageData {
		folderPath := GetImagePath(bundle.Name)
		filePath := path.Join(folderPath, bundle.MakeFileName((*w).GetSuffix()))

		delErr := deleteFile(filePath)

		if delErr != nil {
			return delErr
		}
	}

	return nil
}

// Attempts to roll back any writes that already occrred in the case of an error
func RollBackWrites(data *ImageOutputData) error {
	for _, f := range data.SizeFormats {
		folderPath := GetImagePath(f.Filename)
		filePath := path.Join(folderPath, f.Filename)

		delErr := deleteFile(filePath)

		if delErr != nil {
			return delErr
		}
	}

	return nil
}

// Simple file deletion function.
func deleteFile(filePath string) error {
	return os.Remove(filePath)
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
	filename := makeName()
	writeData := makeJpegBundle(filename)

	imgDat, imageErr := makeJpegData(imageBytes, "original")

	if imageErr != nil {
		return nil, imageErr
	}

	writeData.addNewSize(imgDat.AsImageData())
	writeData.addNewSize(makeThumbnailJpegData(imgDat, "thumb").AsImageData())

	outputData, writeErr := writeFiles(writeData)

	if writeErr != nil {
		return nil, writeErr
	}

	return outputData, nil
}

func processNewPngImage(imageBytes []byte) (*ImageOutputData, error) {
	filename := makeName()
	writeData := makePngBundle(filename)

	imgDat, imageErr := makePngData(imageBytes, "original")

	if imageErr != nil {
		return nil, imageErr
	}

	writeData.addNewSize(imgDat.AsImageData())
	writeData.addNewSize(makeThumbnailPngData(imgDat, "thumb").AsImageData())

	outputData, writeErr := writeFiles(writeData)

	if writeErr != nil {
		return nil, writeErr
	}

	return outputData, nil
}

// TODO provide the ability to convert from bmp to png or jpg
func processNewGifImage(imageBytes []byte) (*ImageOutputData, error) {
	filename := makeName()
	writeData := makeGifBundle(filename)

	imgDat, imageErr := makeGifData(imageBytes, "original")

	if imageErr != nil {
		return nil, imageErr
	}

	writeData.addNewSize(imgDat.AsImageData())
	writeData.addNewSize(makeThumbnailGifData(imgDat, "thumb").AsImageData())

	outputData, writeErr := writeFiles(writeData)

	if writeErr != nil {
		return nil, writeErr
	}

	return outputData, nil
}

// TODO provide the ability to convert from bmp to png or jpg
func processNewBmpImage(imageBytes []byte) (*ImageOutputData, error) {
	filename := makeName()
	writeData := makeBmpBundle(filename)

	imgDat, imageErr := makeBmpData(imageBytes, "original")

	if imageErr != nil {
		return nil, imageErr
	}

	writeData.addNewSize(imgDat.AsImageData())
	writeData.addNewSize(makeThumbnailBmpData(imgDat, "thumb").AsImageData())

	outputData, writeErr := writeFiles(writeData)

	if writeErr != nil {
		return nil, writeErr
	}

	return outputData, nil
}

// TODO compress tiff images to jpeg
func processNewTiffImage(imageBytes []byte) (*ImageOutputData, error) {
	filename := makeName()
	writeData := makeTiffBundle(filename)

	imgDat, imageErr := makeTiffData(imageBytes, "original")

	if imageErr != nil {
		return nil, imageErr
	}

	writeData.addNewSize(imgDat.AsImageData())
	writeData.addNewSize(makeThumbnailTiffData(imgDat, "thumb").AsImageData())

	outputData, writeErr := writeFiles(writeData)

	if writeErr != nil {
		return nil, writeErr
	}

	return outputData, nil
}
