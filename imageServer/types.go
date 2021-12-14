package imageServer

import (
	"os"

	"methompson.com/image-microservice/imageServer/constants"
	"methompson.com/image-microservice/imageServer/dbController"
	"methompson.com/image-microservice/imageServer/imageHandler"
)

// Debug mode is only activated when we set the DEBUG_MODE environment
// variable to true
func DebugMode() bool {
	return os.Getenv(constants.DEBUG_MODE) == "true"
}

// AuthTestingMode is only activated we set the AUTH_TESTING_MODE environment
// variable to true. AuthTestingMode allows a user to perform authenticated
// actions without authentication credentials. Especially useful during
// unit and integration tests
func AuthTestingMode() bool {
	return os.Getenv(constants.AUTH_TESTING_MODE) == "true"
}

type EditImageBody struct {
	Id       string    `json:"id" binding:"required"`
	Title    *string   `json:"title"`
	Filename *string   `json:"filename"`
	Tags     *[]string `json:"tags"`
}

func (ebb *EditImageBody) GetImageDocument() dbController.EditImageDocument {
	doc := dbController.EditImageDocument{
		Id:       ebb.Id,
		Title:    ebb.Title,
		Filename: ebb.Filename,
		Tags:     ebb.Tags,
	}

	return doc
}

type EditImageFileBody struct {
	Id        string `json:"id" binding:"required"`
	Private   *bool  `json:"private"`
	Obfuscate *bool  `json:"obfuscate"`
}

func (doc *EditImageFileBody) GetEditImageFileDocument() dbController.EditImageFileDocument {
	imgDoc := dbController.EditImageFileDocument{
		Id: doc.Id,
	}

	if doc.Private != nil {
		imgDoc.ChangePrivate = true
		imgDoc.Private = *doc.Private
	}

	if doc.Obfuscate != nil {
		imgDoc.ChangeObfuscate = true
		imgDoc.Obfuscate = *doc.Obfuscate
	}

	return imgDoc
}

type DeleteImageBody struct {
	Id string `json:"id" binding:"required"`
}

func (dbb *DeleteImageBody) GetImageDocument() dbController.DeleteImageDocument {
	return dbController.DeleteImageDocument{
		Id: dbb.Id,
	}
}

type DeleteImageFileBody struct {
	Id string `json:"id" binding:"required"`
}

func (dbb *DeleteImageFileBody) GetImageFileDocument() dbController.DeleteImageFileDocument {
	return dbController.DeleteImageFileDocument{
		Id: dbb.Id,
	}
}

type AddImageFormData struct {
	Title      string                           `json:"title"`
	Tags       []string                         `json:"tags"`
	Operations []imageHandler.ConversionRequest `json:"operations"`
}

func GetDefaultImageFormMetaData() AddImageFormData {
	return AddImageFormData{
		Title:      "",
		Tags:       make([]string, 0),
		Operations: make([]imageHandler.ConversionRequest, 0),
	}
}
