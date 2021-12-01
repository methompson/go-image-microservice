package imageConversion

/****************************************************************************************
 * ImageOutputData
*****************************************************************************************/

// The eventual data struct that communicates the result of having written files to the
// filesystem. It provides information, like, name, extension and size formats
type ImageOutputData struct {
	Name             string
	OriginalFileName string
	SizeFormats      []*ImageSizeFormat
}

func (iod *ImageOutputData) AddSizeFormat(sf *ImageSizeFormat) {
	iod.SizeFormats = append(iod.SizeFormats, sf)
}

func makeImageOutputData(iw *ImageWriter, formats []*ImageSizeFormat) *ImageOutputData {
	return &ImageOutputData{
		OriginalFileName: iw.OriginalFileName,
		SizeFormats:      formats,
	}
}
