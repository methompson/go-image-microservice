package mongoDbController

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"methompson.com/image-microservice/imageServer/dbController"
	"methompson.com/image-microservice/imageServer/logging"
)

const IMAGE_COLLECTION = "images"
const LOGGING_COLLECTION = "logging"
const USER_COLLECTION = "users"

type MongoDbController struct {
	MongoClient *mongo.Client
	dbName      string
}

// getCollection is a convenience function that performs a function used regularly
// throughout the Mongodbc. It accepts a collectionName string for the
// specific collection you want to retrieve, and returns a collection, context and
// cancel function.
func (mdbc *MongoDbController) getCollection(collectionName string) (*mongo.Collection, context.Context, context.CancelFunc) {
	// Write the hash to the database
	collection := mdbc.MongoClient.Database(mdbc.dbName).Collection(collectionName)
	backCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	return collection, backCtx, cancel
}

func (mdbc *MongoDbController) initImageCollection(dbName string) error {
	db := mdbc.MongoClient.Database(dbName)

	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{"fileName", "locations", "authorId", "dateAdded"},
		"properties": bson.M{
			"fileName": bson.M{
				"bsonType":    "string",
				"description": "fileName must be a string",
			},
			"locations": bson.M{
				"bsonType":    "array",
				"description": "locations must be an array",
			},
			"authorId": bson.M{
				"bsonType":    "string",
				"description": "authorId must be a string",
			},
			"dateAdded": bson.M{
				"bsonType":    "timestamp",
				"description": "dateAdded must be a timestamp",
			},
		},
	}

	colOpts := options.CreateCollection().SetValidator(bson.M{"$jsonSchema": jsonSchema})

	createCollectionErr := db.CreateCollection(context.TODO(), IMAGE_COLLECTION, colOpts)

	if createCollectionErr != nil {
		return dbController.NewDBError(createCollectionErr.Error())
	}

	models := []mongo.IndexModel{
		{
			Keys:    bson.M{"fileName": 1},
			Options: options.Index().SetUnique(true),
		},
	}

	opts := options.CreateIndexes().SetMaxTime(2 * time.Second)

	collection, _, _ := mdbc.getCollection(IMAGE_COLLECTION)
	_, setIndexErr := collection.Indexes().CreateMany(context.TODO(), models, opts)

	if setIndexErr != nil {
		return dbController.NewDBError(setIndexErr.Error())
	}

	return nil
}

func (mdbc *MongoDbController) initLoggingCollection(dbName string) error {
	db := mdbc.MongoClient.Database(dbName)

	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{"timestamp", "type"},
		"properties": bson.M{
			"timestamp": bson.M{
				"bsonType":    "timestamp",
				"description": "timestamp is required and must be a timestamp",
			},
			"type": bson.M{
				"bsonType":    "string",
				"description": "type is required and must be a string",
			},
		},
	}

	colOpts := options.CreateCollection().SetValidator(bson.M{"$jsonSchema": jsonSchema})
	colOpts.SetCapped(true)
	colOpts.SetSizeInBytes(100000)

	createCollectionErr := db.CreateCollection(context.TODO(), LOGGING_COLLECTION, colOpts)

	if createCollectionErr != nil {
		return dbController.NewDBError(createCollectionErr.Error())
	}

	return nil
}

func (mdbc *MongoDbController) InitDatabase() error {
	imageCreationErr := mdbc.initImageCollection(mdbc.dbName)

	if imageCreationErr != nil && !strings.Contains(imageCreationErr.Error(), "Collection already exists") {
		return imageCreationErr
	}

	loggingCreationErr := mdbc.initLoggingCollection(mdbc.dbName)

	if loggingCreationErr != nil && !strings.Contains(loggingCreationErr.Error(), "Collection already exists") {
		return loggingCreationErr
	}

	return nil
}

func (mdbc *MongoDbController) AddImageData(doc *dbController.AddImageDocument) (string, error) {
	return "", errors.New("Unimplemented")
}

func (mdbc *MongoDbController) GetImageDataById(id string) (*dbController.ImageDocument, error) {
	return nil, errors.New("Unimplemented")
}

func (mdbc *MongoDbController) GetImagesData(page int, pagination int) ([]*dbController.ImageDocument, error) {
	return nil, errors.New("Unimplemented")
}

func (mdbc *MongoDbController) EditImageData(doc *dbController.EditImageDocument) error {
	return errors.New("Unimplemented")
}

func (mdbc *MongoDbController) DeleteImageData(doc *dbController.DeleteImageDocument) error {
	return errors.New("Unimplemented")
}

func (mdbc *MongoDbController) AddRequestLog(log *logging.RequestLogData) error {
	collection, backCtx, cancel := mdbc.getCollection(LOGGING_COLLECTION)
	defer cancel()

	insert := bson.M{
		"timestamp":    primitive.Timestamp{T: uint32(log.Timestamp.Unix())},
		"type":         log.Type,
		"clientIP":     log.ClientIP,
		"method":       log.Method,
		"path":         log.Path,
		"protocol":     log.Protocol,
		"statusCode":   log.StatusCode,
		"latency":      log.Latency,
		"userAgent":    log.UserAgent,
		"errorMessage": log.ErrorMessage,
	}

	_, mdbErr := collection.InsertOne(backCtx, insert)

	if mdbErr != nil {
		return dbController.NewDBError(mdbErr.Error())
	}

	return nil
}

func (mdbc *MongoDbController) AddInfoLog(log *logging.InfoLogData) error {
	collection, backCtx, cancel := mdbc.getCollection(LOGGING_COLLECTION)
	defer cancel()

	insert := bson.M{
		"timestamp": primitive.Timestamp{T: uint32(log.Timestamp.Unix())},
		"type":      log.Type,
		"message":   log.Message,
	}

	_, mdbErr := collection.InsertOne(backCtx, insert)

	if mdbErr != nil {
		return dbController.NewDBError(mdbErr.Error())
	}

	return nil
}
