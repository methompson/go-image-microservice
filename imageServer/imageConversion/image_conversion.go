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

// Starting point for receiving a new image from the user. The gin context and ConversionRequests
// are passed to this function to process the data, determine the image type and perform all
// conversion requests.
func ProcessImageFile(ctx *gin.Context, conversionRequests []ConversionRequest) (ImageOutputData, error) {
	ops := makeNewOpArray()
	for _, req := range conversionRequests {
		op, opErr := makeOpFromRequest(req)

		if opErr == nil {
			ops = append(ops, op)
		}
	}

	file, fileHeader, fileErr := ctx.Request.FormFile("image")

	if fileErr != nil {
		return ImageOutputData{}, fileErr
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	fileBytes, fileBytesErr := ioutil.ReadAll(file)
	if fileBytesErr != nil {
		return ImageOutputData{}, fileBytesErr
	}

	originalFileName := fileHeader.Filename
	// fmt.Println(originalFileName)

	if contentType == "image/heic" {
		return processNewHeifImage(fileBytes, originalFileName, ops)
	} else if contentType == "image/jpeg" ||
		contentType == "image/png" ||
		contentType == "image/gif" ||
		contentType == "image/bmp" ||
		contentType == "image/tiff" {
		return processNewImage(fileBytes, originalFileName, ops)
	}

	return ImageOutputData{}, errors.New("invalid image format")
}

// Returns a string to be used as a file name. Currently just uses UUID
func makeRandomName() string {
	return uuid.New().String()
}

// Attempts to roll back any writes that already occrred in the case of an error
func RollBackWrites(data ImageOutputData) error {
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
func processNewHeifImage(imageBytes []byte, originalFileName string, conversionOps []ConversionOp) (ImageOutputData, error) {
	reader := bytes.NewReader(imageBytes)
	exif, err := goheif.ExtractExif(reader)
	if err != nil {
		return ImageOutputData{}, err
	}

	// os.WriteFile("./files/exif.dat", exif, 0644)

	image, err := goheif.Decode(reader)
	if err != nil {
		return ImageOutputData{}, err
	}

	imgDat := makeImageDataFromImage(&image, Jpeg, exifData{ExifData: exif})

	return convertAndWriteImage(imgDat, originalFileName, conversionOps)
}

// Takes an image file and processes the file based on environment or user parameters.
// imageBytes represents a file send to the function. The function confirms the jpeg
// data, parses the file, performs scale operations and save the data to the file system.
func processNewImage(imageBytes []byte, originalFileName string, conversionOps []ConversionOp) (ImageOutputData, error) {
	imgDat, imageErr := makeImageDataFromBytes(imageBytes)

	if imageErr != nil {
		return ImageOutputData{}, imageErr
	}

	return convertAndWriteImage(imgDat, originalFileName, conversionOps)
}

func convertAndWriteImage(imgDat imageData, originalFileName string, conversionOps []ConversionOp) (ImageOutputData, error) {
	iw := MakeImageWriter(originalFileName, imgDat)
	// iw.AddNewOp(makeOriginalOp())

	// The length of conversionOps will be 1 if the only valid operation is a thumbnail
	// operation. We add an original image operation in order to produce a default
	// series of operations.
	if len(conversionOps) == 1 {
		conversionOps = append(conversionOps, makeOriginalOp())
	}

	for _, op := range conversionOps {
		iw.AddNewOp(op)
	}

	output, writeErr := iw.Commit()

	if writeErr != nil {
		return ImageOutputData{}, writeErr
	}

	return output, nil
}
