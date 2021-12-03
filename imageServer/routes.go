package imageServer

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (srv *ImageServer) SetRoutes() {
	srv.GinEngine.GET("/images", srv.GetImagesByFirstPage)
	srv.GinEngine.GET("/images/page/:page", srv.GetImagesByPage)

	srv.GinEngine.POST("/add-image", srv.SetMaxImageUploadSize, srv.TestLoggedIn, srv.PostAddImage)
	// srv.GinEngine.POST("/add-image", srv.SetMaxImageUploadSize, srv.EnsureLoggedIn, srv.PostAddImage)
	srv.GinEngine.POST("/edit-image", srv.PostEditImage)
	srv.GinEngine.POST("/delete-image", srv.PostDeleteImage)
}

func (srv *ImageServer) SetMaxImageUploadSize(ctx *gin.Context) {
	// TODO set an env variable for max body size
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 5<<20)
	ctx.Next()
}

func (srv *ImageServer) TestLoggedIn(ctx *gin.Context) {
	ctx.Set("userRole", "admin")
	ctx.Set("userId", "1234567890")
}

func (srv *ImageServer) EnsureLoggedIn(ctx *gin.Context) {
	requestUser, getTokenErr := srv.GetRequestUserFromHeader(ctx)

	// No Token Error
	if getTokenErr != nil {
		fmt.Println(getTokenErr)
		ctx.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{"error": "invalid token"},
		)
		return
	}

	// Role Error
	if !srv.CanEditImages(requestUser.Role) {
		ctx.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{"error": "not authorized"},
		)
		return
	}

	ctx.Set("userRole", requestUser.Role)
	ctx.Set("userId", requestUser.Token.UID)

	ctx.Next()
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
// * Parses the metadata from the form body
// * Retrieves the image information from the form body
// * Performs any image conversions, including generating a thumbnail
// * Saves all files to the file system
// * Sends a write to the database.
// If either of the persistent storage functions fail (db or fs), an
// undo command will revert either action in order to prevent ghost
// entries from existing.
func (srv *ImageServer) PostAddImage(ctx *gin.Context) {
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
