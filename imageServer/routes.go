package imageServer

import (
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"methompson.com/image-microservice/imageServer/dbController"
)

func (srv *ImageServer) SetRoutes() {
	srv.GinEngine.Use(srv.ParseRequestUserAuth)

	// /image/:imageName serves an image file
	srv.GinEngine.GET("/image/:imageName", srv.GetImageByName)

	// /images/id/:imageId will serve information about an image.
	srv.GinEngine.GET("/images/id/:imageId", srv.TestLoggedIn, srv.GetImagesById)

	// /images and /images/page/:page will serve pagination information about images
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

// The functino determines which source we use to retrieve the authentication token
// If a token is in the header, we parse the header. If there's no token in the header,
// we parse the cookie.
func (srv *ImageServer) ParseRequestUserAuth(ctx *gin.Context) {
	if srv.HasAuthHeader(ctx) {
		srv.ParseRequestUserHeaders(ctx)
	} else {
		srv.ParseRequestUserCookies(ctx)
	}

	ctx.Next()
}

func (srv *ImageServer) ParseRequestUserHeaders(ctx *gin.Context) {
	token, tokenErr := srv.GetAuthorizationHeader(ctx)

	srv.HandleParsedToken(ctx, token, tokenErr)
}

func (srv *ImageServer) ParseRequestUserCookies(ctx *gin.Context) {
	token, tokenErr := srv.GetAuthorizationCookie(ctx)

	srv.HandleParsedToken(ctx, token, tokenErr)
}

func (srv *ImageServer) HandleParsedToken(ctx *gin.Context, token *auth.Token, tokenErr error) {
	if tokenErr != nil {
		return
	}

	ctx.Set("userId", token.UID)

	role, roleErr := srv.GetRoleFromToken(token)

	if roleErr != nil {
		return
	}

	ctx.Set("role", role)
}

func (srv *ImageServer) EnsureLoggedIn(ctx *gin.Context) {
	userId := ctx.GetString("userId")

	// No Token Error
	if len(userId) <= 0 {
		ctx.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{"error": "invalid token"},
		)
		return
	}

	// Role Error
	role := ctx.GetString("userRole")

	if len(role) <= 0 && !srv.CanEditImages(role) {
		ctx.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{"error": "not authorized"},
		)
		return
	}

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

func (srv *ImageServer) GetImageByName(ctx *gin.Context) {
	filepath, imgDoc, fileErr := srv.ImageController.GetImageByName(ctx)

	// TODO abstract if images can be seen
	if fileErr != nil || !canViewImage(ctx, imgDoc) {
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"error": "file not found",
			},
		)
		return
	}

	// ctx.Header("Content-Description", "File Transfer")
	// ctx.Header("Content-Transfer-Encoding", "binary")
	// ctx.Header("Content-Disposition", "attachment; filename="+imgDoc.Filename)
	ctx.Header("Content-Type", imgDoc.GetMimeType())
	ctx.File(filepath)
}

// Determines if the image requires privileges and whether the requesting user
// has those privileges
func canViewImage(ctx *gin.Context, imgDoc dbController.ImageFileDocument) bool {
	token := ctx.GetString("userId")

	// If not private, return true
	if !imgDoc.Private {
		return true
	}

	// If we get here, the image is private. We just determine if there's a token
	return len(token) > 0
}

func (srv *ImageServer) GetImagesById(ctx *gin.Context) {
	id := ctx.Param("imageId")

	// Not sure this will ever happen
	if len(id) == 0 {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid id",
			},
		)

		return
	}

	doc, err := srv.ImageController.GetImageDataById(id)

	if err != nil {
		ctx.JSON(
			http.StatusNotFound,
			gin.H{},
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		doc.GetMap(),
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
