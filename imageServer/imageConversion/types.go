package imageConversion

import "image"

type ImageData struct {
	ImageData *image.Image
	Suffix    string
}

// func (id *ImageData) MakeFileName(extension string) string {
// 	return id.FileName + "@" + id.Suffix + "." + extension
// }

type ExifData struct {
	ExifData []byte
}

func (exif *ExifData) MakeSizeData() []byte {
	markerlen := 2 + len(exif.ExifData)

	// The size of the marker is represented as 2 bytes, so we have to convert the
	// length into two bytes to place into the exif marker
	sizeByte1 := uint8(markerlen >> 8)
	sizeByte2 := uint8(markerlen & 0xff)

	exifMarker := []byte{0xff, 0xe1, sizeByte1, sizeByte2}

	return exifMarker
}

func (exif *ExifData) MakeFileData() []byte {
	data := exif.MakeSizeData()
	data = append(data, exif.ExifData...)

	return data
}
