package imageServer

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"methompson.com/image-microservice/imageServer/dbController"
	iconv "methompson.com/image-microservice/imageServer/imageConversion"
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
	scaleRequests := make([]*iconv.ScaleRequest, 0)
	scaleRequests = append(scaleRequests, &iconv.ScaleRequest{
		LongestSide: 1000,
		Suffix:      "web",
	})
	scaleRequests = append(scaleRequests, &iconv.ScaleRequest{
		LongestSide: 1000,
		Suffix:      "wide-web",
		ByWidth:     true,
	})

	output, conversionErr := iconv.ProcessImageFile(ctx, scaleRequests)

	if conversionErr != nil {
		return conversionErr
	}

	// iconv.RollBackWrites(output)

	fmt.Println(output.Name)

	return nil
}
