package imageServer

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	jpeg "image/jpeg"
	png "image/png"
	"io/ioutil"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nfnt/resize"

	"methompson.com/image-microservice/imageServer/dbController"
	"methompson.com/image-microservice/imageServer/logging"
)

type ImageController struct {
	DBController *dbController.DatabaseController
	Loggers      []*logging.ImageLogger
}

func InitController(dbc *dbController.DatabaseController) ImageController {
	ic := ImageController{
		DBController: dbc,
		Loggers:      make([]*logging.ImageLogger, 0),
	}

	return ic
}

func (ic *ImageController) AddLogger(logger *logging.ImageLogger) {
	ic.Loggers = append(ic.Loggers, logger)
}

func (ic *ImageController) AddImageFile(ctx *gin.Context) error {
	file, fileHandler, fileErr := ctx.Request.FormFile("image")

	if fileErr != nil {
		fmt.Println("fileErr")
		return fileErr
	}
	defer file.Close()

	fmt.Println(fileHandler.Filename)
	fmt.Println(fileHandler.Size)
	fmt.Println(fileHandler.Header)

	contentType := fileHandler.Header.Get("Content-Type")

	fmt.Println(contentType)

	fileBytes, fileBytesErr := ioutil.ReadAll(file)

	if fileBytesErr != nil {
		fmt.Println("fileBytesErr")
		return fileBytesErr
	}

	originalImage, format, imageErr := image.Decode(bytes.NewReader(fileBytes))

	if imageErr != nil {
		fmt.Println("imageErr")
		return imageErr
	}

	filename := makeName()

	imagesToEncode := make([]*ImageData, 0)
	imagesToEncode = append(imagesToEncode, &ImageData{
		ImageData: &originalImage,
		FileName:  filename,
		Suffix:    "original",
		Format:    format,
	})

	thumb := makeThumbnail(&originalImage)

	imagesToEncode = append(imagesToEncode, &ImageData{
		ImageData: thumb,
		FileName:  filename,
		Suffix:    "thumb",
		Format:    format,
	})

	// saveFile(&originalImage, "files/"+filename+"@original.png", format)
	// saveFile(thumb, "files/"+filename+"@thumb.png", format)

	for _, img := range imagesToEncode {
		saveFile(img, "./files")
	}

	fmt.Println("image format: " + format)

	return nil
}

type ImageData struct {
	ImageData *image.Image
	FileName  string
	Suffix    string
	Format    string
}

func (id *ImageData) MakeFileName() string {
	var extension string = ""
	switch id.Format {
	case "png":
		extension = ".png"
	case "gif":
		extension = ".gif"
	case "bmp":
		extension = ".bmp"
	case "jpeg":
		extension = ".jpg"
	}

	return id.FileName + "@" + id.Suffix + extension
}

func makeName() string {
	return uuid.New().String()
}

func makeThumbnail(img *image.Image) *image.Image {
	var thumb = resize.Thumbnail(128, 128, *img, resize.Lanczos3)
	return &thumb
}

func saveFile(imgData *ImageData, folderPath string) error {
	switch imgData.Format {
	case "png":
		return savePngFile(imgData, folderPath)
	case "bmp":
		return saveBmpFile(imgData, folderPath)
	case "gif":
		return saveGifFile(imgData, folderPath)
	default:
		return saveJpgFile(imgData, folderPath)
	}
}

func saveJpgFile(imgData *ImageData, folderPath string) error {
	buf := new(bytes.Buffer)

	opt := jpeg.Options{
		Quality: 75,
	}

	encodeErr := jpeg.Encode(buf, *imgData.ImageData, &opt)

	if encodeErr != nil {
		return encodeErr
	}

	return writeFileData(imgData, folderPath, buf)
}

func savePngFile(imgData *ImageData, folderPath string) error {
	enc := png.Encoder{
		CompressionLevel: png.BestCompression,
	}

	buf := new(bytes.Buffer)

	encodeErr := enc.Encode(buf, *imgData.ImageData)

	if encodeErr != nil {
		return encodeErr
	}

	return writeFileData(imgData, folderPath, buf)
}

func saveGifFile(imgData *ImageData, folderPath string) error { return nil }

func saveBmpFile(imgData *ImageData, folderPath string) error { return nil }

func writeFileData(imgData *ImageData, folderPath string, buffer *bytes.Buffer) error {
	filePath := path.Join(folderPath, imgData.MakeFileName())

	writeErr := os.WriteFile(filePath, buffer.Bytes(), 0644)

	if writeErr != nil {
		fmt.Println(writeErr.Error())
		return writeErr
	}

	return nil
}
