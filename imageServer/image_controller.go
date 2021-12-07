package imageServer

import (
	"fmt"
	"path"
	"time"

	"github.com/gin-gonic/gin"

	"methompson.com/image-microservice/imageServer/dbController"
	"methompson.com/image-microservice/imageServer/imageConversion"
	"methompson.com/image-microservice/imageServer/logging"
)

type ImageController struct {
	DBController *dbController.DatabaseController
	Loggers      []*logging.ImageLogger
}

func InitController(dbc *dbController.DatabaseController) ImageController {
	ic := ImageController{
		DBController: dbc,
		Loggers:      make([]*logging.ImageLogger, 0),
	}

	return ic
}

func (ic *ImageController) AddLogger(logger *logging.ImageLogger) {
	ic.Loggers = append(ic.Loggers, logger)
}

func (ic *ImageController) AddImageFile(ctx *gin.Context) error {
	metaStr := ctx.PostForm("meta")
	imageFormData := parseAddImageFormString(metaStr)

	output, conversionErr := imageConversion.ProcessImageFile(ctx, imageFormData.Operations)

	if conversionErr != nil {
		return conversionErr
	}

	addImgDoc := dbController.AddImageDocument{
		Title:       imageFormData.Title,
		Tags:        imageFormData.Tags,
		IdName:      output.IdName,
		FileName:    output.OriginalFileName,
		SizeFormats: output.SizeFormats,
		AuthorId:    ctx.GetString("userId"),
		DateAdded:   time.Now(),
	}

	fmt.Println(output.OriginalFileName)
	fmt.Println(addImgDoc.AuthorId)

	_, addImageErr := (*ic.DBController).AddImageData(addImgDoc)

	if addImageErr != nil {
		// TODO Rollback database writes
		imageConversion.RollBackWrites(output)
		return addImageErr
	}

	return nil
}

func (ic *ImageController) GetImages(page, paginationNum int) (docs []dbController.ImageDocument, err error) {
	var _pagination int
	if paginationNum <= 0 {
		_pagination = 50
	} else {
		_pagination = paginationNum
	}
	docs, err = (*ic.DBController).GetImagesData(page, _pagination)
	return
}

func (ic *ImageController) GetImageByName(ctx *gin.Context) (filepath string, imgDoc dbController.ImageFileDocument, err error) {
	name := ctx.Param("imageName")

	fmt.Println(name)

	img, imgErr := (*ic.DBController).GetImageByName(name)

	if imgErr != nil {
		err = imgErr
		return
	}

	filename := img.Filename
	// Make a file path from the file name
	filepath = path.Join(imageConversion.GetImagePath(filename), filename)
	imgDoc = img

	return
}

func (ic *ImageController) GetImageDataById(id string) (doc dbController.ImageDocument, err error) {
	doc, err = (*ic.DBController).GetImageDataById(id)

	return
}
