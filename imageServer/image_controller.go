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
	convRequests := make([]*iconv.ConversionRequest, 0)
	convRequests = append(convRequests, &iconv.ConversionRequest{
		Suffix:   "thumb",
		ResizeOp: "thumbnail",
	})

	var longestSide uint = 1000
	convRequests = append(convRequests, &iconv.ConversionRequest{
		LongestSide: &longestSide,
		Suffix:      "web",
		ResizeOp:    "scale",
	})
	convRequests = append(convRequests, &iconv.ConversionRequest{
		LongestSide: &longestSide,
		Suffix:      "wide-web",
		ResizeOp:    "scalebywidth",
	})

	output, conversionErr := iconv.ProcessImageFile(ctx, convRequests)

	if conversionErr != nil {
		return conversionErr
	}

	// iconv.RollBackWrites(output)

	fmt.Println(output.Name)

	return nil
}
