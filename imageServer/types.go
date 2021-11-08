package imageServer

import (
	"os"
	"time"

	"methompson.com/image-microservice/imageServer/constants"
	"methompson.com/image-microservice/imageServer/dbController"
)

func DebugMode() bool {
	return os.Getenv(constants.GIN_MODE) != "release"
}

type AuthorizationHeader struct {
	Token string `header:"authorization" binding:"required"`
}

type AddImageBody struct {
	Title          string    `json:"title" binding:"required"`
	FileName       string    `json:"fileName" binding:"required"`
	Tags           *[]string `json:"tags"`
	AuthorId       string    `json:"authorId" binding:"required"`
	DateAdded      int       `json:"dateAdded" binding:"required"`
	UpdateAuthorId *string   `json:"updateAuthorId"`
	DateUpdated    *int      `json:"dateUpdated"`
}

func (abb *AddImageBody) GetBlogDocument() *dbController.AddImageDocument {
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
	Id             string    `json:"id" binding:"required"`
	Title          *string   `json:"title"`
	FileName       *string   `json:"fileName"`
	Tags           *[]string `json:"tags"`
	AuthorId       *string   `json:"authorId"`
	DateAdded      *int      `json:"dateAdded"`
	UpdateAuthorId *string   `json:"updateAuthorId"`
	DateUpdated    *int      `json:"dateUpdated"`
}

func (ebb *EditImageBody) GetBlogDocument() *dbController.EditImageDocument {
	var dateAdded *time.Time
	var dateUpdated *time.Time

	if ebb.DateAdded != nil {
		t := time.Unix(int64(*ebb.DateAdded), 0)
		dateAdded = &t
	}

	if ebb.DateUpdated != nil {
		t := time.Unix(int64(*ebb.DateUpdated), 0)
		dateUpdated = &t
	}

	doc := dbController.EditImageDocument{
		Id:             ebb.Id,
		Title:          ebb.Title,
		FileName:       ebb.FileName,
		Tags:           ebb.Tags,
		AuthorId:       ebb.AuthorId,
		DateAdded:      dateAdded,
		UpdateAuthorId: ebb.UpdateAuthorId,
		DateUpdated:    dateUpdated,
	}

	return &doc
}

type DeleteImageBody struct {
	Id string `json:"id" binding:"required"`
}

func (dbb *DeleteImageBody) GetBlogDocument() *dbController.DeleteImageDocument {
	doc := dbController.DeleteImageDocument{
		Id: dbb.Id,
	}

	return &doc
}
