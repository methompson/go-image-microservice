package imageServer

import (
	"fmt"
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

	// imageConversion.RollBackWrites(output)

	fmt.Println(output.OriginalFileName)
	fmt.Println(addImgDoc.AuthorId)

	_, addImageErr := (*ic.DBController).AddImageData(addImgDoc)

	if addImageErr != nil {
		return addImageErr
	}

	return nil
}
