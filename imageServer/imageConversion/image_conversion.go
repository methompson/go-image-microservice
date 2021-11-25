package imageConversion

import (
	"bytes"
	"errors"
	"fmt"
	_ "image/gif"
	"io/ioutil"

	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jdeng/goheif"
)

func ProcessImageFile(ctx *gin.Context, scaleRequests []*ScaleRequest) (*ImageOutputData, error) {
	file, fileHandler, fileErr := ctx.Request.FormFile("image")

	if fileErr != nil {
		return nil, fileErr
	}
	defer file.Close()

	originalFileName := fileHandler.Filename
	fmt.Println(originalFileName)

	contentType := fileHandler.Header.Get("Content-Type")

	fileBytes, fileBytesErr := ioutil.ReadAll(file)

	if fileBytesErr != nil {
		return nil, fileBytesErr
	}

	switch contentType {
	case "image/heic":
		return processNewHeifImage(fileBytes, scaleRequests)
	case "image/jpeg":
		return processNewImage(fileBytes, scaleRequests)
	case "image/png":
		return processNewImage(fileBytes, scaleRequests)
	case "image/gif":
		return processNewImage(fileBytes, scaleRequests)
	case "image/bmp":
		return processNewImage(fileBytes, scaleRequests)
	case "image/tiff":
		return processNewImage(fileBytes, scaleRequests)
	default:
		return nil, errors.New("invalid image format")
	}
}

// Returns a string to be used as a file name. Currently just uses UUID
func makeRandomName() string {
	return uuid.New().String()
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
// * Get an *image.Image struct
// * Get the exif
// Then we pass the above two points to the encode Jpeg function.
func processNewHeifImage(imageBytes []byte, scaleRequests []*ScaleRequest) (*ImageOutputData, error) {
	reader := bytes.NewReader(imageBytes)
	exif, err := goheif.ExtractExif(reader)
	if err != nil {
		return nil, err
	}

	image, err := goheif.Decode(reader)
	if err != nil {
		return nil, err
	}

	imgDat := makeImageDataFromImageWithExif(&image, Jpeg, "original", &exifData{ExifData: exif})
	newBytes, err := (*imgDat).EncodeImage()

	if err != nil {
		return nil, err
	}

	return processNewImage(newBytes, scaleRequests)
}

// Takes an image file and processes the file based on environment or user parameters.
// imageBytes represents a file send to the function. The function confirms the jpeg
// data, parses the file, performs scale operations and save the data to the file system.
func processNewImage(imageBytes []byte, scaleRequests []*ScaleRequest) (*ImageOutputData, error) {
	imgDat, imageErr := makeImageDataFromBytes(imageBytes, "original")

	if imageErr != nil {
		return nil, imageErr
	}

	iw := MakeImageWriter()
	iw.AddNewFile(imgDat)
	iw.AddNewFile(imgDat.MakeThumbnail())

	output, writeErr := iw.Commit()

	if writeErr != nil {
		return nil, writeErr
	}

	return output, nil
}
