package imageServer

import (
	"log"

	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
)

func MakeAndStartServer() {
	envErr := checkEnvVariables()

	if envErr != nil {
		log.Fatal("Error with environment variables")
	}

	// blogServer, srvErr := makeServer()

	// if srvErr != nil {
	// 	log.Fatal("Error making server")
	// }

	// // We run this after creating a server, but before setting routes. Any
	// // route set BEFORE this won't actually use this.
	// if !DebugMode() {
	// 	errs := configureReleaseLogging(blogServer)

	// 	if len(errs) > 0 {
	// 		for _, err := range errs {
	// 			print(err.Error() + "\n")
	// 		}
	// 	}
	// 	addLogging(blogServer)
	// 	addCorsMiddleware(blogServer)

	// 	addRecovery(blogServer)
	// }

	// blogServer.SetRoutes()

	// blogServer.StartServer()
}

// TODO check environment variables
func checkEnvVariables() error {
	return nil
}

type BlogServer struct {
	FirebaseApp    *firebase.App
	BlogController BlogController
	GinEngine      *gin.Engine
}
