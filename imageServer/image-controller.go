package imageServer

import (
	"methompson.com/image-microservice/imageServer/logging"
)

type BlogController struct {
	// DBController *dbController.DatabaseController
	Loggers []*logging.ImageLogger
}
