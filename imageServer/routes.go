package imageServer

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (srv *ImageServer) SetRoutes() {
	srv.GinEngine.GET("/image", srv.GetImagesByFirstPage)
	srv.GinEngine.GET("/image/page/:page", srv.GetImagesByPage)

	srv.GinEngine.POST("/add-image", srv.PostAddImage)
	srv.GinEngine.POST("/edit-image", srv.PostEditImage)
	srv.GinEngine.POST("/delete-image", srv.PostDeleteImage)
}

func (srv *ImageServer) GetImages(ctx *gin.Context) {
	ctx.JSON(
		http.StatusOK,
		gin.H{},
	)
}

func (srv *ImageServer) GetImagesByFirstPage(ctx *gin.Context) {
	ctx.JSON(
		http.StatusOK,
		gin.H{},
	)
}

func (srv *ImageServer) GetImagesByPage(ctx *gin.Context) {
	ctx.JSON(
		http.StatusOK,
		gin.H{},
	)
}

func (srv *ImageServer) PostAddImage(ctx *gin.Context) {
	image, imageErr := ctx.FormFile("image")

	if imageErr != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			gin.H{
				"error": "error retrieving blog posts",
			},
		)
		return
	}

	fmt.Println(image.Filename)

	ctx.JSON(
		http.StatusOK,
		gin.H{},
	)
}

func (srv *ImageServer) PostEditImage(ctx *gin.Context) {
	ctx.JSON(
		http.StatusOK,
		gin.H{},
	)
}

func (srv *ImageServer) PostDeleteImage(ctx *gin.Context) {
	ctx.JSON(
		http.StatusOK,
		gin.H{},
	)
}
