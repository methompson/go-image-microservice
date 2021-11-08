package dbController

import (
	"time"
)

type UserDataDocument struct {
	Id    string
	UID   string
	Name  string
	Role  string
	Email string
}

type AddImageDocument struct {
	Title     string
	FileName  string
	Tags      *[]string
	AuthorId  string
	DateAdded time.Time
}

type ImageDocument struct {
	Id             string
	Title          string
	FileName       string
	Tags           []string
	Locations      []*ImageLocation
	Author         string
	AuthorId       string
	DateAdded      time.Time
	UpdateAuthor   string
	UpdateAuthorId string
	DateUpdated    time.Time
}

type ImageLocation struct {
	SizeType  string
	Url       string
	FileSize  int
	ImageSize *ImageSize
}

type ImageSize struct {
	Width  int
	Height int
}

func (bd *ImageDocument) GetMap() *map[string]interface{} {
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

	locations := make([]*map[string]interface{}, 0)

	for _, loc := range bd.Locations {
		locVal := make(map[string]interface{})
		locVal["sizeType"] = loc.SizeType
		locVal["url"] = loc.Url
		locVal["fileSize"] = loc.FileSize

		size := make(map[string]interface{})
		size["width"] = loc.ImageSize.Width
		size["height"] = loc.ImageSize.Height
		locVal["imageSize"] = &size

		locations = append(locations, &locVal)
	}

	m["locations"] = locations

	return &m
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
