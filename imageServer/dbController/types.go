package dbController

import (
	"strings"
	"time"

	"methompson.com/image-microservice/imageServer/imageHandler"
)

type UserDataDocument struct {
	Id    string
	UID   string
	Name  string
	Role  string
	Email string
}

// A struct representing a new image
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
	SizeFormats []imageHandler.ImageSizeFormat
	AuthorId    string
	DateAdded   time.Time
}

// An image file result for when a user is accessing JUST an image file
type ImageFileDocument struct {
	Id          string
	ImageId     string
	ImageIdName string
	Filename    string
	FormatName  string
	ImageSize   imageHandler.ImageSize
	FileSize    int
	Private     bool
	ImageType   imageHandler.ImageType
}

func (ifd ImageFileDocument) GetMimeType() string {
	var mimeType string

	switch ifd.ImageType {
	case imageHandler.Jpeg:
		mimeType = "image/jpeg"
	case imageHandler.Png:
		mimeType = "image/png"
	case imageHandler.Gif:
		mimeType = "image/gif"
	case imageHandler.Bmp:
		mimeType = "image/bmp"
	case imageHandler.Tiff:
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
	Filename   string
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
	m["fileName"] = bd.Filename
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
	Id       string
	Title    *string
	FileName *string
	Tags     *[]string
}

type EditImageFileDocument struct {
	Id              string
	Private         bool
	ChangePrivate   bool
	Obfuscate       bool
	ChangeObfuscate bool
	NewName         string
}

func (eifd *EditImageFileDocument) ChangesExist() bool {
	return eifd.ChangePrivate || eifd.ChangeObfuscate
}

type EditImageFileResult struct {
	OldName string
	NewName string
}

type DeleteImageDocument struct {
	Id string
}

type DeleteImageFileDocument struct {
	Id string
}

type SortType int8

const (
	Name SortType = iota
	NameReverse
	DateAdded
	DateAddedReverse
)

// A struct for sorting image documents when getting multiple values
type SortImageFilter struct {
	Sortby   SortType
	SearchBy string
}

func MakeSortImageFilter(sortByStr string) SortImageFilter {
	var sortBy SortType
	switch strings.ToLower(sortByStr) {
	case "name":
		sortBy = Name
	case "namereverse":
		sortBy = NameReverse
	case "dateaddedreverse":
		sortBy = DateAddedReverse
	case "dateadded":
	default:
		sortBy = DateAdded
	}

	return SortImageFilter{
		Sortby:   sortBy,
		SearchBy: "",
	}
}
