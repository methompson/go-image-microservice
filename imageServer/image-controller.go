package imageServer

import (
	"methompson.com/image-microservice/imageServer/dbController"
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
