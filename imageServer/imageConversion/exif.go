package imageConversion

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

	"image/jpeg"

	"github.com/jdeng/goheif"
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

func newWriterExif(writer io.Writer, exif []byte) (io.Writer, error) {
	writerSkipper := &writerSkipper{writer, 2}

	// jpeg file signature. jpeg file formats start with FF D8
	// soi = start of image.
	soi := []byte{0xff, 0xd8}
	if _, err := writer.Write(soi); err != nil {
		return nil, err
	}

	if exif != nil {
		markerlen := 2 + len(exif)

		fmt.Printf("heif exif length: %v\n", len(exif))
		fmt.Printf("heif marker length: %v\n", markerlen)

		// The size of the marker is represented as 2 bytes, so we have to convert the
		// length into two bytes to place into the exif marker
		sizeByte1 := uint8(markerlen >> 8)
		sizeByte2 := uint8(markerlen & 0xff)

		// The exif marker starts with FF E1, then the size.
		exifMarker := []byte{0xff, 0xe1, sizeByte1, sizeByte2}

		if _, err := writer.Write(exifMarker); err != nil {
			return nil, err
		}

		// Here, we write the existing exif data to the writer.
		if _, err := writer.Write(exif); err != nil {
			return nil, err
		}
	}

	return writerSkipper, nil
}

func encodeJpegWithExif(imageData *ImageData, exifData ExifData) ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer, _ := newWriterExif(buffer, exifData.ExifData)

	opt := &jpeg.Options{
		Quality: 75,
	}

	err := jpeg.Encode(writer, *imageData.ImageData, opt)

	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func ConvertHeifToJpg(fileBytes []byte) ([]byte, error) {
	reader := bytes.NewReader(fileBytes)
	exif, err := goheif.ExtractExif(reader)
	if err != nil {
		return nil, err
	}

	path := path.Join("./files", "heif_exif.bin")
	os.WriteFile(path, exif, 0644)

	image, err := goheif.Decode(reader)
	if err != nil {
		return nil, err
	}

	buffer := new(bytes.Buffer)
	writer, _ := newWriterExif(buffer, exif)

	opt := &jpeg.Options{
		Quality: 75,
	}

	err = jpeg.Encode(writer, image, opt)

	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
