package dbController

import (
	"time"

	"methompson.com/image-microservice/imageServer/imageConversion"
)

type UserDataDocument struct {
	Id    string
	UID   string
	Name  string
	Role  string
	Email string
}

// A struct representing image about a new image
// Title is a user provided description of the image.
// FileName is the original file name of the file when uploaded
// IdName is a UUID string that is provided to images when they are not obfuscated. it connects images to the ImageDocument
// Tags is a user provided collection of descriptor strings for the image
// SizeFormats represents the actual image files and metadata about each image, such as size and resolution
// AuthorId is the id of the uploader of the image
// DateAdded is the date when the image was uploaded
type AddImageDocument struct {
	Title       string
	FileName    string
	IdName      string
	Tags        []string
	SizeFormats []imageConversion.ImageSizeFormat
	AuthorId    string
	DateAdded   time.Time
}

// An image file result for when a user is accessing JUST an image file
type ImageFileDocument struct {
	Id         string
	Filename   string
	FormatName string
	ImageSize  imageConversion.ImageSize
	FileSize   int
	Private    bool
	ImageType  imageConversion.ImageType
}

func (ifd ImageFileDocument) GetMimeType() string {
	var mimeType string

	switch ifd.ImageType {
	case imageConversion.Jpeg:
		mimeType = "image/jpeg"
	case imageConversion.Png:
		mimeType = "image/png"
	case imageConversion.Gif:
		mimeType = "image/gif"
	case imageConversion.Bmp:
		mimeType = "image/bmp"
	case imageConversion.Tiff:
		mimeType = "image/tiff"
	default:
		mimeType = "application/octet-stream"
	}

	return mimeType
}

func (ifd ImageFileDocument) GetMap() map[string]interface{} {
	m := make(map[string]interface{})

	m["id"] = ifd.Id
	m["filename"] = ifd.Filename
	m["fileSize"] = ifd.FileSize
	m["private"] = ifd.Private
	m["formatName"] = ifd.FormatName
	m["imageSize"] = ifd.ImageSize.GetMap()
	m["imageType"] = ifd.GetMimeType()

	return m
}

type ImageDocument struct {
	Id         string
	Title      string
	FileName   string
	IdName     string
	Tags       []string
	ImageFiles []ImageFileDocument
	Author     string
	AuthorId   string
	DateAdded  time.Time
}

func (bd *ImageDocument) GetMap() map[string]interface{} {
	m := make(map[string]interface{})

	m["id"] = bd.Id

	m["title"] = bd.Title
	m["fileName"] = bd.FileName
	m["author"] = bd.Author
	m["authorId"] = bd.AuthorId
	m["dateAdded"] = bd.DateAdded.Unix()

	if bd.Tags != nil {
		m["tags"] = bd.Tags
	} else {
		m["tags"] = make([]string, 0)
	}

	imageFiles := make([]map[string]interface{}, 0)

	for _, format := range bd.ImageFiles {
		imageFiles = append(imageFiles, format.GetMap())
	}

	m["imageFiles"] = imageFiles

	return m
}

type EditImageDocument struct {
	Id             string
	Title          *string
	FileName       *string
	Tags           *[]string
	AuthorId       *string
	DateAdded      *time.Time
	UpdateAuthorId *string
	DateUpdated    *time.Time
}

type DeleteImageDocument struct {
	Id string
}

type DeleteImageFileDocument struct {
	Id string
}
