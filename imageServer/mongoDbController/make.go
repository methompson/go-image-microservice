package mongoDbController

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"methompson.com/image-microservice/imageServer/constants"
	"methompson.com/image-microservice/imageServer/dbController"
)

func checkEnvVariables() error {
	mongoDbUrl := os.Getenv(constants.MONGO_DB_URL)
	if len(mongoDbUrl) == 0 {
		msg := "MONGO_DB_URL environment variable is required"
		return NewEnvironmentVariableError(msg)
	}

	mongoDbUser := os.Getenv(constants.MONGO_DB_USERNAME)
	if len(mongoDbUser) == 0 {
		msg := "MONGO_DB_USERNAME environment variable is required"
		return NewEnvironmentVariableError(msg)
	}

	mongoDbPass := os.Getenv(constants.MONGO_DB_PASSWORD)
	if len(mongoDbPass) == 0 {
		msg := "MONGO_DB_PASSWORD environment variable is required"
		return NewEnvironmentVariableError(msg)
	}

	return nil
}

// setupMongoDbClient constructs a MongoDB connection URL based on environment
// variables and attempts to connect to the URL. The resulting mongo.Client
// object is returned, and an error is returned.
func setupMongoDbClient() (*mongo.Client, error) {
	envErr := checkEnvVariables()

	if envErr != nil {
		return nil, envErr
	}

	mongoDbUrl := os.Getenv("MONGO_DB_URL")
	mongoDbUser := os.Getenv("MONGO_DB_USERNAME")
	mongoDbPass := os.Getenv("MONGO_DB_PASSWORD")

	mongoDbFullUrl := fmt.Sprintf("mongodb+srv://%v:%v@%v", mongoDbUser, mongoDbPass, mongoDbUrl)
	clientOptions := options.Client().
		ApplyURI(mongoDbFullUrl)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, mdbErr := mongo.Connect(ctx, clientOptions)

	if mdbErr != nil {
		err := fmt.Sprint("Error connecting: ", mdbErr.Error())
		return client, dbController.NewDBError(err)
	}

	return client, nil
}

// The MakeMongoDbController gets a MongoDB client object from
// setupMongoDbClient, then wraps it up in a MongoDbController object along
// with the database name.
func MakeMongoDbController(dbName string) (*MongoDbController, error) {
	client, clientErr := setupMongoDbClient()

	if clientErr != nil {
		return nil, clientErr
	}

	return &MongoDbController{client, dbName}, nil
}
