package imageConversion

import (
	"image"
)

type ImageSize struct {
	Width  int
	Height int
}

func (is ImageSize) GetMap() map[string]interface{} {
	size := make(map[string]interface{})

	size["width"] = is.Width
	size["height"] = is.Height

	return size
}

func GetImageSize(imgData *image.Image) ImageSize {
	bounds := (*imgData).Bounds()

	return ImageSize{
		Width:  bounds.Max.X,
		Height: bounds.Max.Y,
	}
}

// Representation of an actual image that is saved in the file system
// Type is a string description of the type of image. e.g. "thumbnail", "web", "original"
// Filename is the actual file name on the filesystem. The filename is used for accessing the image using a GET command
// ImageSize is an ImageSize struct describing the height and width of the image
// FileSize is the size of the image file in bytes.
// Private is a flag representing whether this image is accessible publicly or not
type ImageSizeFormat struct {
	FormatName string
	Filename   string
	ImageSize  ImageSize
	FileSize   int
	Private    bool
	ImageType  ImageType
}

func (isf ImageSizeFormat) GetMap() map[string]interface{} {
	m := make(map[string]interface{})

	m["filename"] = isf.Filename
	m["fileSize"] = isf.FileSize
	m["private"] = isf.Private
	m["formatName"] = isf.FormatName
	m["imageSize"] = isf.ImageSize.GetMap()

	return m
}

func MakeImageSizeFormat(filename string, fileSize int, imageSize ImageSize, imgOp ConversionOp, imgType ImageType) ImageSizeFormat {
	return ImageSizeFormat{
		FormatName: imgOp.Suffix,
		Filename:   filename,
		ImageSize:  imageSize,
		FileSize:   fileSize,
		Private:    imgOp.Private,
		ImageType:  imgType,
	}
}

/****************************************************************************************
 * ImageOutputData
*****************************************************************************************/

// The eventual data struct that communicates the result of having written files to the
// filesystem. It provides information, like, name, extension and size formats
type ImageConversionResult struct {
	IdName           string
	OriginalFileName string
	SizeFormats      []ImageSizeFormat
}

func (iod *ImageConversionResult) AddSizeFormat(sf ImageSizeFormat) {
	iod.SizeFormats = append(iod.SizeFormats, sf)
}

func makeImageConversionResult(iw *ImageWriter, idName string, formats []ImageSizeFormat) ImageConversionResult {
	return ImageConversionResult{
		IdName:           idName,
		OriginalFileName: iw.OriginalFileName,
		SizeFormats:      formats,
	}
}
