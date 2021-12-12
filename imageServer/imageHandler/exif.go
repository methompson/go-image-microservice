package imageHandler

import (
	"encoding/binary"
	"io"
)

type Orientation uint8

const (
	Horizontal Orientation = 1
	Rotate180  Orientation = 3
	RotateCW   Orientation = 6
	RotateCCW  Orientation = 8
)

func getOrientation(value byte) Orientation {
	switch value {
	case 3:
		return Rotate180
	case 6:
		return RotateCW
	case 8:
		return RotateCCW
	default:
		return Horizontal
	}
}

/****************************************************************************************
 * exifData
*****************************************************************************************/

// Representation of EXIF data, plus convenience functions for generating data
// needed for encoding jpeg info.
type exifData struct {
	ExifData []byte
}

func (exif *exifData) hasData() bool { return (exif.ExifData != nil && len(exif.ExifData) > 0) }

// Generates APP1 Marker bytes and File sizes.
func (exif *exifData) makeSizeData() []byte {
	markerlen := 2 + len(exif.ExifData)

	// The size of the marker is represented as 2 bytes, so we have to convert the
	// length into two bytes to place into the exif marker
	sizeByte1 := uint8(markerlen >> 8)
	sizeByte2 := uint8(markerlen & 0xff)

	// FF E1 are the first two bytes for the exif Marker
	exifMarker := []byte{0xff, 0xe1, sizeByte1, sizeByte2}

	return exifMarker
}

// Generates the marker, file size and appends the exif data.
func (exif *exifData) makeFileData() []byte {
	data := exif.makeSizeData()
	data = append(data, exif.ExifData...)

	return data
}

// This function reviews the exif binary data, looks for a specific byte order
// extracts the information after the metadata and compares it to the spec to
// determine if the image is rotated.
// Typical IFD ordering for an EXIF tag:
// TTTT | ffff | NNNNNNNN | DDDDDDDD
// TTTT (2 bytes) is the tag
// ffff (2 bytes) is the format
// NNNNNNNN (4 bytes) is the number of components
// DDDDDDDD (4 bytes) contains a data value or offset data value
//Example:  01 12 00 03 00 00 00 01 00 06
// 01 12 is the tag (orientation)
// 00 03 is the format (Unsigned short, 2 bytes)
// 00 00 00 01 is the amount of compeonts (1)
// 00 06 is the actual value. In this case, 6 or rotate Clockwise
// Returns an orientation enum
func (exif *exifData) isImageRotated() Orientation {
	bytesLength := len(exif.ExifData)
	for i := 0; i < bytesLength-1; i++ {
		byte1 := exif.ExifData[i]
		byte2 := exif.ExifData[i+1]

		// We check that byte1 & byte2 == 0x0112 and that there are 16 more
		// bytes after 0x0112
		if byte1 == 0x01 && byte2 == 0x12 && bytesLength-i > 8 {
			orientationByte := exif.ExifData[i+9]
			return getOrientation(orientationByte)
		}
	}

	return Horizontal
}

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
func newWriterExif(writer io.Writer, exif exifData) (io.Writer, error) {
	writerSkipper := &writerSkipper{writer, 2}

	// jpeg file signature. jpeg file formats start with FF D8
	// soi = start of image.
	soi := []byte{0xff, 0xd8}
	if _, err := writer.Write(soi); err != nil {
		return nil, err
	}

	if exif.hasData() {
		exifData := exif.makeFileData()

		if _, err := writer.Write(exifData); err != nil {
			return nil, err
		}
	}

	return writerSkipper, nil
}

// TODO find jpeg begining bytes or EXIF end bytes
// FF C0 or something like that
func extractJpegExif(imageBytes []byte) exifData {
	bytesLength := len(imageBytes)
	if bytesLength < 2 {
		return exifData{}
	}

	// Check for jpeg magic bytes
	if imageBytes[0] != 0xff || imageBytes[1] != 0xd8 {
		return exifData{}
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
		return exifData{}
	}

	end := start + length

	if len(imageBytes) < end {
		return exifData{}
	}

	// We have to add 2 to go past the length bytes
	exif := imageBytes[start+2 : end]

	return exifData{
		ExifData: exif,
	}
}
