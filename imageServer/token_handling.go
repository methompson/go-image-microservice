package imageServer

import (
	"context"
	"errors"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

func (srv *ImageServer) HasAuthHeader(ctx *gin.Context) bool {
	return len(ctx.GetHeader("authorization")) > 0
}

func (srv *ImageServer) ValidateIdToken(tokenStr string) (*auth.Token, error) {
	ctx := context.Background()
	client, clientErr := srv.FirebaseApp.Auth(ctx)

	if clientErr != nil {
		return nil, clientErr
	}

	token, tokenErr := client.VerifyIDToken(ctx, tokenStr)

	if tokenErr != nil {
		return nil, tokenErr
	}

	return token, nil
}

func (srv *ImageServer) GetAuthorizationCookie(ctx *gin.Context) (*auth.Token, error) {
	tokenCookie, tokenCookieErr := ctx.Cookie("authorization")

	if tokenCookieErr != nil {
		return nil, tokenCookieErr
	}

	return srv.ValidateIdToken(tokenCookie)
}

func (srv *ImageServer) GetAuthorizationHeader(ctx *gin.Context) (*auth.Token, error) {
	tokenStr := ctx.GetHeader("authorization")

	return srv.ValidateIdToken(tokenStr)
}

func (srv *ImageServer) GetRoleFromToken(token *auth.Token) (string, error) {
	roleInt := token.Claims["role"]

	role, ok := roleInt.(string)

	if !ok {
		return "", errors.New("role is not a string")
	}

	return role, nil
}
