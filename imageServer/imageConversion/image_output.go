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

type ImageSizeFormat struct {
	Filename  string
	ImageSize ImageSize
	FileSize  int
	Private   bool
}

func MakeImageSizeFormat(filename string, fileSize int, imageSize ImageSize, private bool) ImageSizeFormat {
	return ImageSizeFormat{
		Filename:  filename,
		ImageSize: imageSize,
		FileSize:  fileSize,
		Private:   private,
	}
}

/****************************************************************************************
 * ImageOutputData
*****************************************************************************************/

// The eventual data struct that communicates the result of having written files to the
// filesystem. It provides information, like, name, extension and size formats
type ImageOutputData struct {
	Name             string
	OriginalFileName string
	SizeFormats      []ImageSizeFormat
}

func (iod *ImageOutputData) AddSizeFormat(sf ImageSizeFormat) {
	iod.SizeFormats = append(iod.SizeFormats, sf)
}

func makeImageOutputData(iw *ImageWriter, formats []ImageSizeFormat) ImageOutputData {
	return ImageOutputData{
		OriginalFileName: iw.OriginalFileName,
		SizeFormats:      formats,
	}
}
