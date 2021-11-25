package imageConversion

import (
	"os"
	"path"
)

// Name is the file name of the files to be written. Extension is the file type
// extension to be appended to the end of the file name. imagesToCommit are the
// individual image data structs that will eventually be written.
type ImageWriter struct {
	Name           string
	imagesToCommit map[string]*imageData
}

func (iw *ImageWriter) makeFileName(imgDat *imageData) string {
	dat := *imgDat
	return iw.Name + "@" + imgDat.Suffix + "." + dat.GetExtension()
}

func (iw *ImageWriter) AddNewFile(imgDat *imageData) {
	iw.imagesToCommit[imgDat.Suffix] = imgDat
}

func (iw *ImageWriter) Commit() (*ImageOutputData, error) {
	sizeFormats := make([]*ImageSizeFormat, 0)
	for _, imgDat := range iw.imagesToCommit {
		size, sizeErr := iw.writeNewFile(imgDat)

		if sizeErr != nil {
			iw.rollback()
			return nil, sizeErr
		}

		sizeFormats = append(sizeFormats, size)
	}

	return makeImageOutputData(iw, sizeFormats), nil
}

// Rollback image writes
func (iw *ImageWriter) rollback() {
	for _, imgDat := range iw.imagesToCommit {
		filename := iw.makeFileName(imgDat)
		folderPath := GetImagePath(filename)
		filePath := path.Join(folderPath, filename)

		deleteFile(filePath)
	}
}

func (iw *ImageWriter) writeNewFile(imgData *imageData) (*ImageSizeFormat, error) {
	size, name, writeError := iw.writeFile(imgData)

	if writeError != nil {
		return nil, writeError
	}

	imgSize := MakeImageSizeFormat(name, size, (*imgData).GetImageSize())

	return imgSize, nil
}

// Takes an imageData instance and writes the contents to the proper folder. Returns
// two values and an error. The first return value is the file size of the resultant
// file write. The second return value is the name of the file.
func (iw *ImageWriter) writeFile(imgData *imageData) (filesize int, filename string, err error) {
	dat := *imgData
	filename = iw.makeFileName(imgData)

	folderPath := GetImagePath(filename)

	folderErr := CheckOrCreateImageFolder(folderPath)

	if folderErr != nil {
		return -1, "", folderErr
	}

	filePath := path.Join(folderPath, filename)
	bytes, encodeErr := dat.EncodeImage()

	if encodeErr != nil {
		return -1, "", encodeErr
	}

	writeErr := os.WriteFile(filePath, bytes, 0644)

	if writeErr != nil {
		return -1, "", writeErr
	}

	return len(bytes), filename, nil
}

func MakeImageWriter() *ImageWriter {
	name := makeRandomName()
	return &ImageWriter{
		Name:           name,
		imagesToCommit: make(map[string]*imageData),
	}
}

/****************************************************************************************
 * ImageOutputData
*****************************************************************************************/

// The eventual data struct that communicates the result of having written files to the
// filesystem. It provides information, like, name, extension and size formats
type ImageOutputData struct {
	Name        string
	SizeFormats []*ImageSizeFormat
}

func (iod *ImageOutputData) AddSizeFormat(sf *ImageSizeFormat) {
	iod.SizeFormats = append(iod.SizeFormats, sf)
}

func makeImageOutputData(iw *ImageWriter, formats []*ImageSizeFormat) *ImageOutputData {
	return &ImageOutputData{
		Name:        iw.Name,
		SizeFormats: formats,
	}
}
