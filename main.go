package main

import (
	"github.com/joho/godotenv"

	"methompson.com/image-microservice/imageServer"
)

func main() {
	godotenv.Load()

	imageServer.MakeAndStartServer()
}
