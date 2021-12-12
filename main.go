package main

import (
	"log"

	"github.com/joho/godotenv"

	"methompson.com/image-microservice/imageServer"
	"methompson.com/image-microservice/imageServer/imageHandler"
)

func main() {
	godotenv.Load()

	imgFolderErr := imageHandler.CheckOrCreateImageFolder(imageHandler.GetImageRootPath())

	if imgFolderErr != nil {
		log.Fatal(imgFolderErr.Error())
	}

	imageServer.MakeAndStartServer()
}
