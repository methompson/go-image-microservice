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

type ImageDocument struct {
	Id             string
	Title          string
	FileName       string
	IdName         string
	Tags           []string
	SizeFormats    []imageConversion.ImageSizeFormat
	Author         string
	AuthorId       string
	DateAdded      time.Time
	UpdateAuthor   string
	UpdateAuthorId string
	DateUpdated    time.Time
}

func (bd *ImageDocument) GetMap() map[string]interface{} {
	m := make(map[string]interface{})

	m["id"] = bd.Id

	m["title"] = bd.Title
	m["fileName"] = bd.FileName
	m["author"] = bd.Author
	m["authorId"] = bd.AuthorId
	m["dateAdded"] = bd.DateAdded.Unix()
	m["updateAuthor"] = bd.UpdateAuthor
	m["updateAuthorId"] = bd.UpdateAuthorId
	m["dateUpdated"] = bd.DateUpdated.Unix()

	if bd.Tags != nil {
		m["tags"] = bd.Tags
	} else {
		m["tags"] = make([]string, 0)
	}

	sizeFormats := make([]map[string]interface{}, 0)

	for _, loc := range bd.SizeFormats {
		locVal := make(map[string]interface{})

		locVal["filename"] = loc.Filename
		locVal["fileSize"] = loc.FileSize
		locVal["private"] = loc.Private

		size := make(map[string]interface{})
		size["width"] = loc.ImageSize.Width
		size["height"] = loc.ImageSize.Height
		locVal["imageSize"] = size

		sizeFormats = append(sizeFormats, locVal)
	}

	m["sizeFormats"] = sizeFormats

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
