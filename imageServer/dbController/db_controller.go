package dbController

import (
	"methompson.com/image-microservice/imageServer/logging"
)

type DatabaseController interface {
	InitDatabase() error

	AddImageData(doc AddImageDocument) (id string, err error)

	GetImageByName(id string) (ImageFileDocument, error)

	GetImageDataById(id string) (ImageDocument, error)
	GetImagesData(page int, pagination int) ([]ImageDocument, error)

	EditImageData(doc EditImageDocument) error
	DeleteImage(doc DeleteImageDocument) error
	DeleteImageFile(doc DeleteImageFileDocument) error

	AddRequestLog(log logging.RequestLogData) error
	AddInfoLog(log logging.InfoLogData) error
}
