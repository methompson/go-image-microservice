package mongoDbController

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"methompson.com/image-microservice/imageServer/dbController"
	"methompson.com/image-microservice/imageServer/logging"
)

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

func (mdbc *MongoDbController) initImageFileCollection(dbName string) error {
	db := mdbc.MongoClient.Database(dbName)

	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{
			"formatName",
			"filename",
			"imageSize",
			"fileSize",
			"private",
		},
		"properties": bson.M{
			"formatName": bson.M{
				"bsonType":    "string",
				"description": "formatName must be a string",
			},
			"filename": bson.M{
				"bsonType":    "string",
				"description": "fileName must be a string",
			},
			"imageSize": bson.M{
				"bsonType":    "object",
				"description": "imageSize must be an array of image size objects",
				"properties": bson.M{
					"width": bson.M{
						"bsonType":    "int",
						"description": "width must be an int",
					},
					"height": bson.M{
						"bsonType":    "int",
						"description": "height must be an int",
					},
				},
			},
			"fileSize": bson.M{
				"bsonType":    "int",
				"description": "fileSize must be an int",
			},
			"private": bson.M{
				"bsonType":    "bool",
				"description": "private must be a bool",
			},
		},
	}

	colOpts := options.CreateCollection().SetValidator(bson.M{"$jsonSchema": jsonSchema})

	createCollectionErr := db.CreateCollection(context.TODO(), IMAGE_FILE_COLLECTION, colOpts)

	if createCollectionErr != nil {
		fmt.Println(createCollectionErr.Error())
		return dbController.NewDBError(createCollectionErr.Error())
	}

	return nil
}

func (mdbc *MongoDbController) initImageCollection(dbName string) error {
	db := mdbc.MongoClient.Database(dbName)

	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{
			"title",
			"fileName",
			"idName",
			"tags",
			"imageIds",
			"authorId",
			"dateAdded",
		},
		"properties": bson.M{
			"title": bson.M{
				"bsonType":    "string",
				"description": "title must be a string",
			},
			"fileName": bson.M{
				"bsonType":    "string",
				"description": "fileName must be a string",
			},
			"idName": bson.M{
				"bsonType":    "string",
				"description": "idName must be a string",
			},
			"tags": bson.M{
				"bsonType":    "array",
				"description": "tags must be an array",
				"items": bson.M{
					"bsonType":    "string",
					"description": "Tag Items must be string",
				},
			},
			"imageIds": bson.M{
				"bsonType":    "array",
				"description": "imageIds must be an array",
				"items": bson.M{
					"bsonType":    "string",
					"description": "imageIds Items must be string",
				},
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

	imageFileCreationErr := mdbc.initImageFileCollection(mdbc.dbName)

	if imageFileCreationErr != nil && !strings.Contains(imageFileCreationErr.Error(), "Collection already exists") {
		return imageFileCreationErr
	}

	loggingCreationErr := mdbc.initLoggingCollection(mdbc.dbName)

	if loggingCreationErr != nil && !strings.Contains(loggingCreationErr.Error(), "Collection already exists") {
		return loggingCreationErr
	}

	return nil
}

func (mdbc *MongoDbController) AddImageData(doc dbController.AddImageDocument) (string, error) {
	imgFileCollection, ifCtx, ifCancel := mdbc.getCollection(IMAGE_FILE_COLLECTION)
	defer ifCancel()

	docs := make([]interface{}, 0)

	for _, img := range doc.SizeFormats {
		docs = append(docs, bson.M{
			"formatName": img.FormatName,
			"filename":   img.Filename,
			"imageSize": bson.M{
				"width":  img.ImageSize.Width,
				"height": img.ImageSize.Height,
			},
			"fileSize": img.FileSize,
			"private":  img.Private,
		})
	}

	fileInsertResult, fileInsertErr := imgFileCollection.InsertMany(ifCtx, docs)

	if fileInsertErr != nil {
		return "", dbController.NewDBError(fileInsertErr.Error())
	}

	imageIds := make([]string, 0)

	for _, result := range fileInsertResult.InsertedIDs {
		id, idOk := result.(primitive.ObjectID)
		if !idOk {
			fmt.Println("Id not OK")
			continue
		}

		// fmt.Println("id: " + id.Hex())
		imageIds = append(imageIds, id.Hex())
	}

	imgCollection, imgCtx, imgCancel := mdbc.getCollection(IMAGE_COLLECTION)
	defer imgCancel()

	col := bson.M{
		"title":     doc.Title,
		"fileName":  doc.FileName,
		"idName":    doc.IdName,
		"tags":      doc.Tags,
		"imageIds":  imageIds,
		"authorId":  doc.AuthorId,
		"dateAdded": primitive.Timestamp{T: uint32(doc.DateAdded.Unix())},
	}

	colInsertResult, colInsertErr := imgCollection.InsertOne(imgCtx, col)

	if colInsertErr != nil {
		return "", dbController.NewDBError(colInsertErr.Error())
	}

	objectId, idOk := colInsertResult.InsertedID.(primitive.ObjectID)

	if !idOk {
		return "", dbController.NewDBError("invalid id returned by database")
	}

	return objectId.Hex(), nil
}

func (mdbc *MongoDbController) GetImageDataById(id string) (imgDoc dbController.ImageDocument, err error) {
	return imgDoc, errors.New("Unimplemented")
}

func (mdbc *MongoDbController) GetImagesData(page int, pagination int) ([]dbController.ImageDocument, error) {
	return nil, errors.New("Unimplemented")
}

func (mdbc *MongoDbController) EditImageData(doc dbController.EditImageDocument) error {
	return errors.New("Unimplemented")
}

func (mdbc *MongoDbController) DeleteImageData(doc dbController.DeleteImageDocument) error {
	return errors.New("Unimplemented")
}

func (mdbc *MongoDbController) AddRequestLog(log logging.RequestLogData) error {
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

func (mdbc *MongoDbController) AddInfoLog(log logging.InfoLogData) error {
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
