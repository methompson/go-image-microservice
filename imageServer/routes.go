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

// POST /add-image
// Adding an Image to the server. It accomplishes the task as follows:
// * Checks the user's authentication
// * Parses the metadata from the form body
// * Retrieves the image information from the form body
// * Performs any image conversions, including generating a thumbnail
// * Saves all files to the file system
// * Sends a write to the database.
// If either of the persistent storage functions fail (db or fs), an
// undo command will revert either action in order to prevent ghost
// entries from existing.
func (srv *ImageServer) PostAddImage(ctx *gin.Context) {
	metaStr := ctx.PostForm("meta")
	meta := srv.ParseFormMetadata(metaStr)
	fmt.Println(meta.Private)

	fileSaveErr := srv.ImageController.AddImageFile(ctx)

	if fileSaveErr != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			gin.H{
				"error": fileSaveErr.Error(),
			},
		)
		return
	}

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
