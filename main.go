package main

import (
	"log"

	"github.com/joho/godotenv"

	"methompson.com/image-microservice/imageServer"
	"methompson.com/image-microservice/imageServer/imageConversion"
)

func main() {
	godotenv.Load()

	imgFolderErr := imageConversion.CheckOrCreateImageFolder(imageConversion.GetImageRootPath())

	if imgFolderErr != nil {
		log.Fatal(imgFolderErr.Error())
	}

	imageServer.MakeAndStartServer()
}
