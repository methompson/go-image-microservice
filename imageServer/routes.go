package imageServer

import (
	"net/http"
	"strconv"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"

	"methompson.com/image-microservice/imageServer/dbController"
)

func (srv *ImageServer) SetRoutes() {
	srv.GinEngine.Use(srv.ParseRequestUserAuth)
	srv.GinEngine.Use(srv.SetMaxImageUploadSize)

	// /image/:imageName serves an image file
	srv.GinEngine.GET("/image/:imageName", srv.GetImageByName)

	// /images/id/:imageId will serve information about an image.
	srv.GinEngine.GET("/image/id/:imageId", srv.EnsureLoggedIn, srv.GetImageById)

	// /images and /images/page/:page will serve pagination information about images
	srv.GinEngine.GET("/images", srv.GetImagesByFirstPage)
	srv.GinEngine.GET("/images/page/:page", srv.GetImagesByPage)

	srv.GinEngine.POST("/add-image", srv.EnsureLoggedIn, srv.PostAddImage)
	srv.GinEngine.POST("/edit-image-file", srv.EnsureLoggedIn, srv.PostEditImageFile)
	srv.GinEngine.POST("/delete-image", srv.EnsureLoggedIn, srv.PostDeleteImage)
	srv.GinEngine.POST("/delete-image-file", srv.EnsureLoggedIn, srv.PostDeleteImageFile)
}

func (srv *ImageServer) SetMaxImageUploadSize(ctx *gin.Context) {
	// TODO set an env variable for max body size
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 5<<20)
}

func (srv *ImageServer) TestLoggedIn(ctx *gin.Context) {
	ctx.Set("userRole", "admin")
	ctx.Set("userId", "1234567890")
}

// The function determines which source we use to retrieve the authentication token
// If a token is in the header, we parse the header. If there's no token in the header,
// we parse the cookie. If we have the AUTH_TESTING_MODE variable set to true, we
// are in a testing mode and we just use a pass-through to get our tokens.
func (srv *ImageServer) ParseRequestUserAuth(ctx *gin.Context) {
	if AuthTestingMode() {
		srv.TestLoggedIn(ctx)
	} else if srv.HasAuthHeader(ctx) {
		srv.ParseRequestUserHeaders(ctx)
	} else {
		srv.ParseRequestUserCookies(ctx)
	}
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

// GET

func (srv *ImageServer) GetImagesByFirstPage(ctx *gin.Context) {
	srv.GetImages(ctx, 1)
}

// We won't worry about throwing errors in this function. We'll just use the default
// value of 1 (first page) if the value for page is mangled or garbage
func (srv *ImageServer) GetImagesByPage(ctx *gin.Context) {
	pageNum, pageNumErr := strconv.Atoi(ctx.Param("page"))

	if pageNumErr != nil {
		pageNum = 1
	}

	srv.GetImages(ctx, pageNum)
}

// TODO start defining filters that users can pass via query parameters
func (srv *ImageServer) GetImages(ctx *gin.Context, page int) {
	pagination := ctx.Query("pagination")
	sortBy := ctx.Query("sortBy")

	paginationNum, paginationNumErr := strconv.Atoi(pagination)
	if paginationNumErr != nil {
		paginationNum = -1
	}

	showPrivate := userLoggedIn(ctx)

	images, err := srv.ImageController.GetImages(page, paginationNum, sortBy, showPrivate)

	if err != nil {
		handleControllerErrors(ctx, err)
		return
	}

	output := make([]map[string]interface{}, 0)

	for _, val := range images {
		output = append(output, val.GetMap())
	}

	ctx.JSON(
		http.StatusOK,
		output,
	)
}

func (srv *ImageServer) GetImageByName(ctx *gin.Context) {
	filepath, imgDoc, err := srv.ImageController.GetImageByName(ctx)

	if err != nil {
		handleControllerErrors(ctx, err)
		return
	}

	if !canViewImage(ctx, imgDoc) {
		ctx.AbortWithStatusJSON(
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

func (srv *ImageServer) GetImageById(ctx *gin.Context) {
	showPrivate := userLoggedIn(ctx)

	doc, err := srv.ImageController.GetImageDataById(ctx, showPrivate)

	if err != nil {
		handleControllerErrors(ctx, err)
		return
	}

	ctx.JSON(
		http.StatusOK,
		doc.GetMap(),
	)
}

// POST

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
	err := srv.ImageController.AddImageFile(ctx)

	if err != nil {
		handleControllerErrors(ctx, err)

		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{},
	)
}

func (srv *ImageServer) PostEditImageFile(ctx *gin.Context) {
	var body EditImageFileBody

	if bindJsonErr := ctx.ShouldBindJSON(&body); bindJsonErr != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			gin.H{"error": "missing required values"},
		)
		return
	}

	err := srv.ImageController.EditImageFileDocument(body)

	if err != nil {
		handleControllerErrors(ctx, err)

		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{},
	)
}

func (srv *ImageServer) PostDeleteImage(ctx *gin.Context) {
	// Extract the body
	var body DeleteImageBody

	if bindJsonErr := ctx.ShouldBindJSON(&body); bindJsonErr != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			gin.H{"error": "missing required values"},
		)
		return
	}

	err := srv.ImageController.DeleteImageDocument(body.GetImageDocument())

	if err != nil {
		handleControllerErrors(ctx, err)
		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{},
	)
}

func (srv *ImageServer) PostDeleteImageFile(ctx *gin.Context) {
	// Extract the body
	var body DeleteImageFileBody

	if bindJsonErr := ctx.ShouldBindJSON(&body); bindJsonErr != nil {
		ctx.AbortWithStatusJSON(
			http.StatusBadRequest,
			gin.H{"error": "missing required values"},
		)
		return
	}

	err := srv.ImageController.DeleteImageFileDocument(body.GetImageFileDocument())

	if err != nil {
		handleControllerErrors(ctx, err)
		return
	}

	ctx.JSON(
		http.StatusOK,
		gin.H{},
	)
}

func handleControllerErrors(ctx *gin.Context, err error) {
	var status int
	var message string
	switch err.(type) {
	case dbController.InvalidInputError:
		status = http.StatusBadRequest
		message = "invalid input"
	case dbController.NoResultsError:
		status = http.StatusNotFound
		message = "not found"
	default:
		status = http.StatusInternalServerError
		message = "internal server error"
	}

	ctx.AbortWithStatusJSON(
		status,
		gin.H{
			"error": message,
		},
	)
}

func userLoggedIn(ctx *gin.Context) bool {
	return len(ctx.GetString("userId")) > 0
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
