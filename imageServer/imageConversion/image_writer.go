package imageConversion

import (
	"bytes"
	"encoding/binary"
	"errors"
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
	"github.com/jdeng/goheif"
	_ "github.com/jdeng/goheif"
	"github.com/nfnt/resize"
)

func SaveImageFile(ctx *gin.Context) error {
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

	fileBytes, fileBytesErr := ioutil.ReadAll(file)

	if fileBytesErr != nil {
		fmt.Println("fileBytesErr")
		return fileBytesErr
	}

	switch contentType {
	case "image/heic":
		return processHeifImage(fileBytes)
	case "image/jpeg":
		return processJpegImage(fileBytes)
	case "image/png":
		return processPngImage(fileBytes)
	case "image/gif":
		return processGifImage(fileBytes)
	case "image/bmp":
		return processBmpImage(fileBytes)
	case "image/tiff":
		return processTiffImage(fileBytes)
	default:
		return errors.New("invalid image format")
	}

	// originalImage, format, imageErr := image.Decode(bytes.NewReader(fileBytes))

	// if imageErr != nil {
	// 	fmt.Println("imageErr")
	// 	return imageErr
	// }

	// filename := makeName()

	// imagesToEncode := make([]*ImageData, 0)
	// imagesToEncode = append(imagesToEncode, &ImageData{
	// 	ImageData: &originalImage,
	// 	FileName:  filename,
	// 	Suffix:    "original",
	// 	Format:    format,
	// })

	// thumb := makeThumbnail(&originalImage)

	// imagesToEncode = append(imagesToEncode, &ImageData{
	// 	ImageData: thumb,
	// 	FileName:  filename,
	// 	Suffix:    "thumb",
	// 	Format:    format,
	// })

	// fmt.Println("image format: " + format)
}

func makeName() string {
	return uuid.New().String()
}

func makeThumbnail(img *image.Image) *image.Image {
	var thumb = resize.Thumbnail(128, 128, *img, resize.Lanczos3)
	return &thumb
}

func resizeImage(img *image.Image, longestSide int) *image.Image {
	width := (*img).Bounds().Max.X
	height := (*img).Bounds().Max.Y

	fmt.Printf("width: %v, height: %v\n", width, height)
	return nil
}

func encodeJpegFile(imgData *ImageData, exifData ExifData, folderPath string) ([]byte, error) {
	buf := new(bytes.Buffer)

	// TODO set quality option or make it an environment variable
	encodeErr := jpeg.Encode(buf, *imgData.ImageData, &jpeg.Options{
		Quality: 75,
	})

	if encodeErr != nil {
		return nil, encodeErr
	}

	// filePath := path.Join(folderPath, imgData.MakeFileName("jpg"))

	// return writeFileData(filePath, buf.Bytes())

	return buf.Bytes(), nil
}

func encodePngFile(imgData *ImageData, folderPath string) error {
	enc := png.Encoder{
		CompressionLevel: png.BestCompression,
	}

	buf := new(bytes.Buffer)

	encodeErr := enc.Encode(buf, *imgData.ImageData)

	if encodeErr != nil {
		return encodeErr
	}

	// filePath := path.Join(folderPath, imgData.MakeFileName("png"))

	// return writeFileData(filePath, buf.Bytes())

	return nil
}

func encodeTiffFile() {}
func encodeGifFile()  {}
func encodeBmpFile()  {}

func writeFileData(filePath string, fileBytes []byte) error {
	writeErr := os.WriteFile(filePath, fileBytes, 0644)

	if writeErr != nil {
		fmt.Println(writeErr.Error())
		return writeErr
	}

	return nil
}

func makeNewImages(img *image.Image) []ImageData {
	images := make([]ImageData, 0)

	images = append(images, ImageData{
		Suffix:    "thumb",
		ImageData: makeThumbnail(img),
	})

	images = append(images, ImageData{
		Suffix:    "web",
		ImageData: resizeImage(img, 800),
	})

	return images
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
func processHeifImage(imageBytes []byte) error {
	// newBytes, err := ConvertHeifToJpg(imageBytes)

	reader := bytes.NewReader(imageBytes)
	exif, err := goheif.ExtractExif(reader)
	if err != nil {
		return err
	}

	// path := path.Join("./files", "heif_exif.bin")
	// os.WriteFile(path, exif, 0644)

	image, err := goheif.Decode(reader)
	if err != nil {
		return err
	}

	newBytes, err := encodeJpegWithExif(
		&ImageData{
			ImageData: &image,
			Suffix:    "fullsize",
		}, ExifData{
			ExifData: exif,
		},
	)

	path := path.Join("./files", "image.jpg")
	writeFileData(path, newBytes)

	if err != nil {
		return err
	}

	return processJpegImage(newBytes)
}

func processJpegImage(imageBytes []byte) error {
	originalImage, _, imageErr := image.Decode(bytes.NewReader(imageBytes))

	if imageErr != nil {
		return imageErr
	}

	extractJpegExif(imageBytes)

	makeNewImages(&originalImage)

	// filename := makeName()

	// path := path.Join("./files", filename+".jpg")
	// writeErr := writeFileData(path, imageBytes)

	// if writeErr != nil {
	// 	return writeErr
	// }

	return nil
}

func extractJpegExif(imageBytes []byte) *ExifData {
	if len(imageBytes) < 6 {
		return nil
	}

	// Check for jpeg magic bytes
	if imageBytes[0] != 0xff || imageBytes[1] != 0xd8 {
		return nil
	}

	fmt.Println("Jpeg bytes exist!")

	// Check for exif bytes
	if imageBytes[2] != 0xff && imageBytes[3] != 0xe1 {
		return nil
	}

	fmt.Println("exif bytes exist!")

	lengthSlice := imageBytes[4:6]

	length := int(binary.BigEndian.Uint16(lengthSlice))

	fmt.Printf("jpeg exif reported length: %v\n", length)

	start := 6
	end := start + length - 2

	if len(imageBytes) < end {
		return nil
	}
	// We have to remove 2 to remove the length value
	exif := imageBytes[start:end]

	fmt.Printf("jpeg exif actual length: %v\n", len(exif))

	path := path.Join("./files", "jpeg_exif.bin")
	os.WriteFile(path, exif, 0644)

	return &ExifData{
		ExifData: exif,
	}
}

func processPngImage(imageBytes []byte) error {
	filename := makeName()

	path := path.Join("./files", filename+".png")
	return writeFileData(path, imageBytes)
}

func processGifImage(imageBytes []byte) error {
	filename := makeName()

	path := path.Join("./files", filename+".gif")
	writeErr := writeFileData(path, imageBytes)

	return writeErr
}

func processBmpImage(imageBytes []byte) error {
	filename := makeName()

	path := path.Join("./files", filename+".bmp")
	writeErr := writeFileData(path, imageBytes)

	return writeErr
}

func processTiffImage(imageBytes []byte) error {
	filename := makeName()

	path := path.Join("./files", filename+".tiff")
	writeErr := writeFileData(path, imageBytes)

	return writeErr
}
