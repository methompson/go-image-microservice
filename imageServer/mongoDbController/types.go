package mongoDbController

import (
	"time"

	"methompson.com/image-microservice/imageServer/dbController"
	"methompson.com/image-microservice/imageServer/imageConversion"
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

type ImageFileDocResult struct {
	Id          string                    `bson:"_id"`
	ImageId     string                    `bson:"imageId"`
	ImageIdName string                    `bson:"imageIdName"`
	Filename    string                    `bson:"filename"`
	FormatName  string                    `bson:"formatName"`
	ImageSize   imageConversion.ImageSize `bson:"imageSize"`
	FileSize    int                       `bson:"fileSize"`
	Private     bool                      `bson:"private"`
	ImageType   string                    `bson:"imageType"`
}

func (ifdr ImageFileDocResult) getImageFileDocument() dbController.ImageFileDocument {
	var imgType imageConversion.ImageType
	switch ifdr.ImageType {
	case "jpeg":
		imgType = imageConversion.Jpeg
	case "png":
		imgType = imageConversion.Png
	case "gif":
		imgType = imageConversion.Gif
	case "bmp":
		imgType = imageConversion.Bmp
	case "tiff":
		imgType = imageConversion.Tiff
	default:
		imgType = imageConversion.Same
	}

	return dbController.ImageFileDocument{
		Id:         ifdr.Id,
		Filename:   ifdr.Filename,
		FormatName: ifdr.FormatName,
		ImageSize:  ifdr.ImageSize,
		FileSize:   ifdr.FileSize,
		Private:    ifdr.Private,
		ImageType:  imgType,
	}
}

func (ifdr ImageFileDocResult) GetMap() map[string]interface{} {
	m := make(map[string]interface{})

	m["id"] = ifdr.Id
	m["filename"] = ifdr.Filename
	m["fileSize"] = ifdr.FileSize
	m["private"] = ifdr.Private
	m["formatName"] = ifdr.FormatName
	m["imageSize"] = ifdr.ImageSize.GetMap()
	m["imageType"] = ifdr.ImageType

	return m
}

type ImageDocResult struct {
	Id        string               `bson:"_id"`
	Title     string               `bson:"title"`
	FileName  string               `bson:"fileName"`
	IdName    string               `bson:"idName"`
	Images    []ImageFileDocResult `bson:"images"`
	Tags      []string             `bson:"tags"`
	Author    []UserDocResult      `bson:"author"`
	AuthorId  string               `bson:"authorId"`
	DateAdded time.Time            `bson:"dateAdded"`
}

func (idr *ImageDocResult) GetImageDocument() dbController.ImageDocument {
	var author string = ""
	if len(idr.Author) > 0 {
		author = idr.Author[0].Name
	}

	imageFiles := make([]dbController.ImageFileDocument, 0)

	for _, res := range idr.Images {
		imageFiles = append(imageFiles, res.getImageFileDocument())
	}

	return dbController.ImageDocument{
		Id:         idr.Id,
		Title:      idr.Title,
		FileName:   idr.FileName,
		IdName:     idr.IdName,
		Tags:       idr.Tags,
		ImageFiles: imageFiles,
		Author:     author,
		AuthorId:   idr.AuthorId,
		DateAdded:  idr.DateAdded,
	}
}
