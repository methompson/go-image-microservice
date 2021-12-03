package mongoDbController

import (
	"time"

	"methompson.com/image-microservice/imageServer/dbController"
)

const IMAGE_COLLECTION = "images"
const IMAGE_FILE_COLLECTION = "imageFiles"
const LOGGING_COLLECTION = "logging"
const USER_COLLECTION = "users"

type UserDocResult struct {
	Id    string `bson:"_id"`
	UID   string `bson:"uid"`
	Name  string `bson:"name"`
	Role  string `bson:"role"`
	Email string `bson:"email"`
}

func (udr *UserDocResult) GetUserDataDoc() *dbController.UserDataDocument {
	doc := dbController.UserDataDocument{
		Id:    udr.Id,
		UID:   udr.UID,
		Name:  udr.Name,
		Role:  udr.Role,
		Email: udr.Email,
	}

	return &doc
}

type ImageDocResult struct {
	Id             string          `bson:"_id"`
	Title          string          `bson:"title"`
	FileName       string          `bson:"fileName"`
	Tags           []string        `bson:"tags"`
	Author         []UserDocResult `bson:"author"`
	AuthorId       string          `bson:"authorId"`
	DateAdded      time.Time       `bson:"dateAdded"`
	UpdateAuthor   []UserDocResult `bson:"updateAuthor"`
	UpdateAuthorId string          `bson:"updateAuthorId"`
	DateUpdated    time.Time       `bson:"dateUpdated"`
}

func (idr *ImageDocResult) GetBlogDocument() *dbController.ImageDocument {
	var author string = ""
	if len(idr.Author) > 0 {
		author = idr.Author[0].Name
	}

	var updateAuthor string = ""
	if len(idr.UpdateAuthor) > 0 {
		updateAuthor = idr.Author[0].Name
	}

	doc := dbController.ImageDocument{
		Id:             idr.Id,
		Title:          idr.Title,
		FileName:       idr.FileName,
		Tags:           idr.Tags,
		Author:         author,
		AuthorId:       idr.AuthorId,
		DateAdded:      idr.DateAdded,
		UpdateAuthor:   updateAuthor,
		UpdateAuthorId: idr.UpdateAuthorId,
		DateUpdated:    idr.DateUpdated,
	}

	return &doc
}
