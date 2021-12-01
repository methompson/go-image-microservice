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

func ProcessImageFile(ctx *gin.Context, scaleRequests []*ConversionRequest) (*ImageOutputData, error) {
	file, fileHeader, fileErr := ctx.Request.FormFile("image")

	if fileErr != nil {
		return nil, fileErr
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	fileBytes, fileBytesErr := ioutil.ReadAll(file)
	if fileBytesErr != nil {
		return nil, fileBytesErr
	}

	originalFileName := fileHeader.Filename
	// fmt.Println(originalFileName)

	if contentType == "image/heic" {
		return processNewHeifImage(fileBytes, originalFileName, scaleRequests)
	} else if contentType == "image/jpeg" ||
		contentType == "image/png" ||
		contentType == "image/gif" ||
		contentType == "image/bmp" ||
		contentType == "image/tiff" {
		return processNewImage(fileBytes, originalFileName, scaleRequests)
	}

	return nil, errors.New("invalid image format")
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
func processNewHeifImage(imageBytes []byte, originalFileName string, conversionRequests []*ConversionRequest) (*ImageOutputData, error) {
	reader := bytes.NewReader(imageBytes)
	exif, err := goheif.ExtractExif(reader)
	if err != nil {
		return nil, err
	}

	// os.WriteFile("./files/exif.dat", exif, 0644)

	image, err := goheif.Decode(reader)
	if err != nil {
		return nil, err
	}

	imgDat := makeImageDataFromImage(&image, Jpeg, &exifData{ExifData: exif})

	return convertAndWriteImage(imgDat, originalFileName, conversionRequests)
}

// Takes an image file and processes the file based on environment or user parameters.
// imageBytes represents a file send to the function. The function confirms the jpeg
// data, parses the file, performs scale operations and save the data to the file system.
func processNewImage(imageBytes []byte, originalFileName string, conversionRequests []*ConversionRequest) (*ImageOutputData, error) {
	imgDat, imageErr := makeImageDataFromBytes(imageBytes)

	if imageErr != nil {
		return nil, imageErr
	}

	return convertAndWriteImage(imgDat, originalFileName, conversionRequests)
}

func convertAndWriteImage(imgDat *imageData, originalFileName string, conversionRequests []*ConversionRequest) (*ImageOutputData, error) {
	iw := MakeImageWriter(originalFileName, imgDat)
	iw.AddNewOp(makeOriginalOp())

	for _, req := range conversionRequests {
		op := makeOpFromRequest(req)
		iw.AddNewOp(op)
	}

	output, writeErr := iw.Commit()

	if writeErr != nil {
		return nil, writeErr
	}

	return output, nil
}
