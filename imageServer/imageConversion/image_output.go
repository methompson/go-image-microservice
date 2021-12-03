package imageConversion

import "image"

type ImageSize struct {
	Width  int
	Height int
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
	Type      string
	Filename  string
	ImageSize ImageSize
	FileSize  int
	Private   bool
}

func MakeImageSizeFormat(filename string, fileSize int, imageSize ImageSize, imgOp ConversionOp) ImageSizeFormat {
	return ImageSizeFormat{
		Type:      imgOp.Suffix,
		Filename:  filename,
		ImageSize: imageSize,
		FileSize:  fileSize,
		Private:   imgOp.Private,
	}
}

/****************************************************************************************
 * ImageOutputData
*****************************************************************************************/

// The eventual data struct that communicates the result of having written files to the
// filesystem. It provides information, like, name, extension and size formats
type ImageOutputData struct {
	IdName           string
	OriginalFileName string
	SizeFormats      []ImageSizeFormat
}

func (iod *ImageOutputData) AddSizeFormat(sf ImageSizeFormat) {
	iod.SizeFormats = append(iod.SizeFormats, sf)
}

func makeImageOutputData(iw *ImageWriter, idName string, formats []ImageSizeFormat) ImageOutputData {
	return ImageOutputData{
		IdName:           idName,
		OriginalFileName: iw.OriginalFileName,
		SizeFormats:      formats,
	}
}
