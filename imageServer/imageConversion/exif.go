package imageConversion

import (
	"encoding/binary"
	"io"
)

// Skip Writer for exif writing
type writerSkipper struct {
	writer      io.Writer
	bytesToSkip int
}

// Overrides the Write function to skip the bytes set by bytesToSkip
func (w *writerSkipper) Write(data []byte) (int, error) {
	// If bytesToSkip has been reduced to 0, we call the writer.Write function
	// as-is with no further processing
	if w.bytesToSkip <= 0 {
		return w.writer.Write(data)
	}

	// If the lens of the bytes passed into Write is less than the remaining
	// bytesToSkip value, we decrement the bytesToSkip value and return the
	// datalength and no error
	if dataLen := len(data); dataLen < w.bytesToSkip {
		w.bytesToSkip -= dataLen
		return dataLen, nil
	}

	// If the amount of data exceeds the bytesToSkip, we make a slice, skipping
	// the bytesToSkip value and extending to the end of the slice.
	dataToWrite := data[w.bytesToSkip:]

	// We write this slice to the writer, then reduce bytesToSkip to zero
	if n, err := w.writer.Write(dataToWrite); err == nil {
		n += w.bytesToSkip
		w.bytesToSkip = 0
		return n, nil
	} else {
		return n, err
	}
}

// Makes a new writerSkipper with the exif data defined. Inserts the exif data
// into the buffer, then returns the writer.
func newWriterExif(writer io.Writer, exif *exifData) (io.Writer, error) {
	writerSkipper := &writerSkipper{writer, 2}

	// jpeg file signature. jpeg file formats start with FF D8
	// soi = start of image.
	soi := []byte{0xff, 0xd8}
	if _, err := writer.Write(soi); err != nil {
		return nil, err
	}

	if exif != nil {
		exifData := exif.makeFileData()

		if _, err := writer.Write(exifData); err != nil {
			return nil, err
		}
	}

	return writerSkipper, nil
}

// TODO find jpeg beinging bytes or EXIF end bytes
// FF C0 or something like that
func extractJpegExif(imageBytes []byte) *exifData {
	bytesLength := len(imageBytes)
	if bytesLength < 2 {
		return nil
	}

	// Check for jpeg magic bytes
	if imageBytes[0] != 0xff || imageBytes[1] != 0xd8 {
		return nil
	}

	length := -1
	start := -1

	// Iterate through the file to find the exif bytes
	for i := 0; i < bytesLength-1; i++ {
		byte1 := imageBytes[i]
		byte2 := imageBytes[i+1]

		if byte1 == 0xff && byte2 == 0xe1 {

			lengthSlice := imageBytes[i+2 : i+4]
			length = int(binary.BigEndian.Uint16(lengthSlice))

			start = i + 2
			break
		}
	}

	if length == -1 {
		return nil
	}

	end := start + length

	if len(imageBytes) < end {
		return nil
	}

	// We have to add 2 to go past the length bytes
	exif := imageBytes[start+2 : end]

	return &exifData{
		ExifData: exif,
	}
}
