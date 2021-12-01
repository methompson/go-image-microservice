package imageConversion

import (
	"errors"
	"os"
	"path"
	"sync"
)

// Name is the file name of the files to be written. Extension is the file type
// extension to be appended to the end of the file name. imagesToCommit are the
// individual image data structs that will eventually be written.
type ImageWriter struct {
	OriginalFileName string
	imageOperations  map[string]*ConversionOp
	imageData        *imageData
}

func (iw *ImageWriter) makeFileName(name string, op *ConversionOp) string {
	return name + "@" + op.Suffix + "." + (*iw).GetExtension(op)
}

func (iw *ImageWriter) GetExtension(op *ConversionOp) string {
	var iType ImageType
	if op.EncodeTo != nil {
		iType = *op.EncodeTo
	} else {
		iType = (*iw).imageData.OriginalImageType
	}

	switch iType {
	case Jpeg:
		return "jpg"
	case Png:
		return "png"
	case Gif:
		return "gif"
	case Bmp:
		return "bmp"
	case Tiff:
		return "tiff"
	default:
		return ""
	}
}

func (iw *ImageWriter) AddNewOp(op *ConversionOp) {
	iw.imageOperations[op.Suffix] = op
}

// Commit takes all image sizes defined in imagesToCommit and writes them all to disk.
// Performs all operations asynchronously, but doesn't finish until all operations are
// finished.
func (iw *ImageWriter) Commit() (*ImageOutputData, error) {
	// This will be the end result
	sizeFormats := make([]*ImageSizeFormat, 0)

	// We get the total operation length and create output and error channels to get
	// necessary information from the operations themselves.
	totalOps := len(iw.imageOperations)

	outputChannel := make(chan *ImageSizeFormat, totalOps)
	errorChannel := make(chan error, totalOps)

	// We use a WaitGroup to sync all operations
	var wg sync.WaitGroup

	// The name will be a UUID to minimize potential name collissions
	var name string
	for _, imgOp := range iw.imageOperations {
		if name == "" || imgOp.Obfuscate {
			name = makeRandomName()
		}

		// We have to assign the value of imgOp to a variable so that it's not changed
		// when the next loop iteration occurs. The go routine can wait until a blocking
		// function occurs before it begins. Assigning imgOp to op prevents the value
		// from changing when it goes out of scope.
		op := imgOp

		// We add a new op to the WaitGroup
		wg.Add(1)
		go func() {
			// We defer the execution of Done until this goroutine is finished executing.
			defer wg.Done()

			// We use the syncronous writeNewFile function to actually write the file
			// and pass the return values to the channels.
			sizeF, writeErr := iw.writeNewFile(op, name)
			outputChannel <- sizeF
			errorChannel <- writeErr
		}()
	}

	// We wait until all image write ops are done.
	wg.Wait()

	// We start a collection of errors. If an error DOES occur, we need to collect
	// at least one. Any write error will result in rolling back the operation
	errs := make([]error, 0)

	for range iw.imageOperations {
		size := <-outputChannel
		writeErr := <-errorChannel

		// Here, we collect the errors into the array and continue the for loop.
		// We need to collect all successful operations so that we can roll them
		// all back. That's why we use continue instead of break
		if writeErr != nil {
			errs = append(errs, writeErr)
			continue
		}

		sizeFormats = append(sizeFormats, size)
	}

	if len(errs) > 0 {
		iw.rollback(sizeFormats)
		return nil, errors.New("write error. rolling back operation")
	}

	return makeImageOutputData(iw, sizeFormats), nil
}

// Rollback image writes
func (iw *ImageWriter) rollback(writtenImages []*ImageSizeFormat) []error {
	errs := make([]error, 0)
	for _, imgDat := range writtenImages {
		folderPath := GetImagePath(imgDat.Filename)
		filePath := path.Join(folderPath, imgDat.Filename)

		err := deleteFile(filePath)

		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// Takes an image operation and name, performs the conversion, gets image information
// and returns the ImageSizeFormat for the converted file. Returns an error if there's
// a problem with the write.
func (iw *ImageWriter) writeNewFile(imgOp *ConversionOp, name string) (*ImageSizeFormat, error) {
	dat := *iw.imageData
	filename := iw.makeFileName(name, imgOp)

	folderPath := GetImagePath(filename)

	folderErr := CheckOrCreateImageFolder(folderPath)

	if folderErr != nil {
		return nil, folderErr
	}

	filePath := path.Join(folderPath, filename)
	bytes, imgSize, encodeErr := dat.EncodeImage(imgOp)

	if encodeErr != nil {
		return nil, encodeErr
	}

	writeErr := os.WriteFile(filePath, bytes, 0644)

	if writeErr != nil {
		return nil, writeErr
	}

	imgSizeF := MakeImageSizeFormat(filename, len(bytes), imgSize)

	return imgSizeF, nil
}

func MakeImageWriter(originalFileName string, imgData *imageData) *ImageWriter {
	return &ImageWriter{
		OriginalFileName: originalFileName,
		imageOperations:  make(map[string]*ConversionOp),
		imageData:        imgData,
	}
}
