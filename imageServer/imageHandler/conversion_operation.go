package imageHandler

import (
	"errors"
	"strings"
)

type ConversionRequest struct {
	// string representation of a file type we want to encode to. The following are valid
	// compressTo values:
	// jpeg
	// gif
	// png
	// bmp
	// tiff
	CompressTo string `json:"compressTo"`

	// A string representation of this file's purpose or size. e.g. "web" or
	// "x-large". If Obfuscate is set to false, this value will be added to
	// the end of the filename.
	Suffix string `json:"suffix"`

	// Longest side in the resize operation. Aspect ratios are maintained, so this
	// value only dictates one side.
	LongestSide uint `json:"longestSide"`

	// Whether to keep all filenames the same or randomize the names
	Obfuscate bool `json:"obfuscate"`

	// string representation of what kind of resize operation will take place.
	// The following are valid ResizeOp values and what they do:
	// original     : Maintains the image as-is with no compression or resizing
	// thumbnail    : Resize the image down to a small size, dictated by the THUMBNAIL_SIZE environment variable or 128px by default
	// scale        : Scales the image, setting the longest side to the LongestSide value. This operation maintains the image's aspect ratio
	// scalebywidth : Scales the image so that the width is set to LongestSide. This operation maintains the image's aspect ratio
	ResizeOp string `json:"resizeOp"`

	// Indicates whether this image should be available publicly or privately.
	Private bool `json:"private"`
}

type ImageType int8

const (
	Same ImageType = iota
	Jpeg
	Png
	Gif
	Bmp
	Tiff
)

type ResizeOp int8

const (
	Original ResizeOp = iota
	Thumbnail
	Scale
	ScaleByWidth
)

// The ConversionOp is a blueprint for an image conversion operation.
// We use the original source image and perform operations on that.
// The ConversionOp also defines a format to encode to, in case the user wants to have
// different sizes in different formats.
type ConversionOp struct {
	// The format that this image will be compressed into. This value can be left as
	// Same in order to use the original image's compression format
	CompressTo ImageType

	// A string representation of this file's purpose or size, e.g. "web" or
	// "x-large". If Obfuscate is set to false, this value will be added to
	// the end of the filename.
	Suffix string

	// Longest side in the resize operation. Aspect ratios are maintained, so this
	// value only dictates one side.
	LongestSide uint

	// Resize operation chosen for this conversion operation.
	ResizeOp ResizeOp

	// This option will randomize the file name.
	Obfuscate bool

	// Indicates whether this image should be available publicly or privately
	Private bool
}

// Takes a ConversionRequest struct and returns a ConversionRequest We return an
// error if the user does not explicitly define a resize operation
func makeOpFromRequest(req ConversionRequest) (ConversionOp, error) {
	var resizeOp ResizeOp
	switch strings.ToLower(req.ResizeOp) {
	case "thumbnail":
		resizeOp = Thumbnail
	case "thumb":
		resizeOp = Thumbnail
	case "scale":
		resizeOp = Scale
	case "scalebywidth":
		resizeOp = ScaleByWidth
	case "original":
		resizeOp = Original
	default:
		return ConversionOp{}, errors.New("invalid resize operation")
	}

	var encodeTo ImageType
	if req.CompressTo != "" {
		switch strings.ToLower(req.CompressTo) {
		case "jpeg":
			encodeTo = Jpeg
		case "png":
			encodeTo = Png
		case "gif":
			encodeTo = Gif
		case "bmp":
			encodeTo = Bmp
		case "tiff":
			encodeTo = Tiff
		default:
			encodeTo = Same
		}
	}

	// We reserve the "thumb" suffix to make sure that there is only one thumbnail
	suffix := req.Suffix
	if suffix == "thumb" && resizeOp != Thumbnail {
		suffix = "thumb_"
	}

	// We return an error if the user does not set the value greater than zero
	// and has an Original or Thumbnail resize operation
	if req.LongestSide == 0 && resizeOp != Original && resizeOp != Thumbnail {
		return ConversionOp{}, errors.New("invalid longest side value or operation")
	}

	return ConversionOp{
		Suffix:      suffix,
		CompressTo:  encodeTo,
		LongestSide: req.LongestSide,
		ResizeOp:    resizeOp,
		Obfuscate:   req.Obfuscate,
		Private:     req.Private,
	}, nil
}

func makeOriginalOp() ConversionOp {
	return ConversionOp{
		Suffix:   "original",
		ResizeOp: Original,
	}
}

func makeThumbnailOp() ConversionOp {
	return ConversionOp{
		Suffix:   "thumb",
		ResizeOp: Thumbnail,
	}
}

// Makes an array of Conversion Operations and adds a thumbnail to this array
// If the user chooses to make the thumbnails private, they can add their own
// thumb operation which should override this operation. Same goes for an
// original operation
func makeNewOpArray() []ConversionOp {
	ops := make([]ConversionOp, 0)
	ops = append(ops, makeThumbnailOp())

	return ops
}
