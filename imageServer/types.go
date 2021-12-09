package imageServer

import (
	"os"
	"time"

	"methompson.com/image-microservice/imageServer/constants"
	"methompson.com/image-microservice/imageServer/dbController"
	"methompson.com/image-microservice/imageServer/imageConversion"
)

func DebugMode() bool {
	return os.Getenv(constants.GIN_MODE) != "release"
}

type AddImageBody struct {
	Title          string   `json:"title" binding:"required"`
	FileName       string   `json:"fileName" binding:"required"`
	Tags           []string `json:"tags"`
	AuthorId       string   `json:"authorId" binding:"required"`
	DateAdded      int      `json:"dateAdded" binding:"required"`
	UpdateAuthorId *string  `json:"updateAuthorId"`
	DateUpdated    *int     `json:"dateUpdated"`
}

func (abb *AddImageBody) GetImageDocument() *dbController.AddImageDocument {
	dateAdded := time.Unix(int64(abb.DateAdded), 0)

	doc := dbController.AddImageDocument{
		Title:     abb.Title,
		FileName:  abb.FileName,
		Tags:      abb.Tags,
		AuthorId:  abb.AuthorId,
		DateAdded: dateAdded,
	}

	return &doc
}

type EditImageBody struct {
	Id       string    `json:"id" binding:"required"`
	Title    *string   `json:"title"`
	FileName *string   `json:"fileName"`
	Tags     *[]string `json:"tags"`
}

type EditImageFileBody struct {
}

func (ebb *EditImageBody) GetImageDocument() *dbController.EditImageDocument {
	doc := dbController.EditImageDocument{
		Id:       ebb.Id,
		Title:    ebb.Title,
		FileName: ebb.FileName,
		Tags:     ebb.Tags,
	}

	return &doc
}

type DeleteImageBody struct {
	Id string `json:"id" binding:"required"`
}

func (dbb *DeleteImageBody) GetImageDocument() dbController.DeleteImageDocument {
	doc := dbController.DeleteImageDocument{
		Id: dbb.Id,
	}

	return doc
}

type AddImageFormData struct {
	Title      string                              `json:"title"`
	Tags       []string                            `json:"tags"`
	Operations []imageConversion.ConversionRequest `json:"operations"`
}

func GetDefaultImageFormMetaData() AddImageFormData {
	return AddImageFormData{
		Title:      "",
		Tags:       make([]string, 0),
		Operations: make([]imageConversion.ConversionRequest, 0),
	}
}

type SortType int8

const (
	Name SortType = iota
	NameReverse
	DateAdded
	DateAddedReverse
)

type ImageFilterSort struct {
	SortBy     SortType
	SearchTerm string
}
