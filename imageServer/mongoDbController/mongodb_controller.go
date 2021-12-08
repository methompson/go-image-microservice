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
	"methompson.com/image-microservice/imageServer/imageConversion"
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	return collection, ctx, cancel
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
			"imageType",
		},
		"properties": bson.M{
			"formatName": bson.M{
				"bsonType":    "string",
				"description": "formatName must be a string",
			},
			"imageType": bson.M{
				"bsonType":    "string",
				"description": "imageType must be a string",
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
					"bsonType":    "objectId",
					"description": "imageIds Items must be ObjectId",
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
			Keys:    bson.M{"idName": 1},
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	imgFileCollection := mdbc.MongoClient.Database(mdbc.dbName).Collection(IMAGE_FILE_COLLECTION)
	imgCollection := mdbc.MongoClient.Database(mdbc.dbName).Collection(IMAGE_COLLECTION)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		docs := make([]interface{}, 0)

		for _, img := range doc.SizeFormats {
			var imgType string

			switch img.ImageType {
			case imageConversion.Jpeg:
				imgType = "jpeg"
			case imageConversion.Png:
				imgType = "png"
			case imageConversion.Gif:
				imgType = "gif"
			case imageConversion.Bmp:
				imgType = "bmp"
			case imageConversion.Tiff:
				imgType = "tiff"
			}

			docs = append(docs, bson.M{
				"formatName": img.FormatName,
				"filename":   img.Filename,
				"imageSize": bson.M{
					"width":  img.ImageSize.Width,
					"height": img.ImageSize.Height,
				},
				"fileSize":  img.FileSize,
				"private":   img.Private,
				"imageType": imgType,
			})
		}

		fileInsertResult, fileInsertErr := imgFileCollection.InsertMany(sessCtx, docs)

		if fileInsertErr != nil {
			return "", dbController.NewDBError(fileInsertErr.Error())
		}

		// imageIds := make([]string, 0)
		imageIds := make([]primitive.ObjectID, 0)

		for _, result := range fileInsertResult.InsertedIDs {
			id, idOk := result.(primitive.ObjectID)
			if !idOk {
				fmt.Println("Id not OK")
				continue
			}

			// fmt.Println("id: " + id.Hex())
			// imageIds = append(imageIds, id.Hex())
			imageIds = append(imageIds, id)
		}

		col := bson.M{
			"title":     doc.Title,
			"fileName":  doc.FileName,
			"idName":    doc.IdName,
			"tags":      doc.Tags,
			"imageIds":  imageIds,
			"authorId":  doc.AuthorId,
			"dateAdded": primitive.Timestamp{T: uint32(doc.DateAdded.Unix())},
		}

		colInsertResult, colInsertErr := imgCollection.InsertOne(sessCtx, col)

		if colInsertErr != nil {
			return "", dbController.NewDBError(colInsertErr.Error())
		}

		if objectId, idOk := colInsertResult.InsertedID.(primitive.ObjectID); idOk {
			return objectId.Hex(), nil
		} else {
			return "", dbController.NewDBError("invalid id returned by database")
		}
	}

	session, sessionErr := mdbc.MongoClient.StartSession()

	if sessionErr != nil {
		return "", sessionErr
	}
	defer session.EndSession(ctx)

	result, transErr := session.WithTransaction(ctx, callback)

	if transErr != nil {
		session.AbortTransaction(ctx)
		return "", transErr
	}

	fmt.Printf("result %v\n", result)

	if id, ok := result.(string); ok {
		return id, nil
	} else {
		return "", errors.New("invalid database response")
	}
}

func (mdbc *MongoDbController) GetImageDataAggregationStages() (projectStage, authorLookupStage, imageFileLookupStage bson.D) {
	projectStage = bson.D{
		{
			Key: "$project", Value: bson.M{
				"title":     1,
				"fileName":  1,
				"idName":    1,
				"tags":      1,
				"imageIds":  1,
				"authorId":  1,
				"dateAdded": 1,
			},
		},
	}

	authorLookupStage = bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         USER_COLLECTION,
			"localField":   "authorId",
			"foreignField": "uid",
			"as":           "author",
		},
	}}

	imageFileLookupStage = bson.D{
		{
			Key: "$lookup",
			Value: bson.M{
				"from":         IMAGE_FILE_COLLECTION,
				"localField":   "imageIds",
				"foreignField": "_id",
				"as":           "images",
			},
		},
	}

	return
}

func (mdbc *MongoDbController) GetImageByName(name string) (imgDoc dbController.ImageFileDocument, err error) {
	collection, ctx, cancel := mdbc.getCollection(IMAGE_FILE_COLLECTION)
	defer cancel()

	var result ImageFileDocResult

	findErr := collection.FindOne(ctx, bson.M{
		"filename": name,
	}).Decode(&result)

	if findErr != nil {
		err = findErr
		return
	}

	return result.getImageFileDocument(), nil
}

func (mdbc *MongoDbController) GetImageDataById(id string) (imgDoc dbController.ImageDocument, err error) {
	idObj, idObjErr := primitive.ObjectIDFromHex(id)

	if idObjErr != nil {
		err = dbController.NewInvalidInputError("invalid id")
		return
	}

	matchStage := bson.D{
		{
			Key: "$match",
			Value: bson.M{
				"_id": idObj,
			},
		},
	}

	return mdbc.GetImageDataWithMatcher(matchStage)
}

func (mdbc *MongoDbController) GetImageDataWithMatcher(matchStage bson.D) (imgDoc dbController.ImageDocument, err error) {
	projectStage, authorLookupStage, imageFileLookupStage := mdbc.GetImageDataAggregationStages()

	collection, ctx, cancel := mdbc.getCollection(IMAGE_COLLECTION)
	defer cancel()

	limitStage := bson.D{{
		Key:   "$limit",
		Value: int32(1),
	}}

	cursor, aggErr := collection.Aggregate(ctx, mongo.Pipeline{
		matchStage,
		projectStage,
		limitStage,
		authorLookupStage,
		imageFileLookupStage,
	})

	if aggErr != nil {
		err = dbController.NewDBError("error getting data from database: " + aggErr.Error())
		return
	}

	var results []ImageDocResult

	if allErr := cursor.All(ctx, &results); allErr != nil {
		err = dbController.NewDBError("error parsing results: " + allErr.Error())
		return
	}

	if len(results) < 1 {
		err = dbController.NewNoResultsError("")
		return
	}

	imgDoc = results[0].GetImageDocument()

	return
}

func (mdbc *MongoDbController) GetImagesData(page int, pagination int) (imgDocs []dbController.ImageDocument, err error) {
	collection, ctx, cancel := mdbc.getCollection(IMAGE_COLLECTION)
	defer cancel()

	matchStage := bson.D{{Key: "$match", Value: bson.M{}}}

	projectStage, authorLookupStage, imageFileLookupStage := mdbc.GetImageDataAggregationStages()

	sortStage := bson.D{{
		Key: "$sort",
		Value: bson.M{
			"dateAdded": -1,
		},
	}}

	limitStage := bson.D{{
		Key:   "$limit",
		Value: int32(pagination),
	}}

	skipStage := bson.D{{
		Key:   "$skip",
		Value: int64((page - 1) * pagination),
	}}

	cursor, aggErr := collection.Aggregate(ctx, mongo.Pipeline{
		matchStage,
		projectStage,
		sortStage,
		skipStage,
		limitStage,
		authorLookupStage,
		imageFileLookupStage,
	})

	if aggErr != nil {
		err = aggErr
		return
	}

	var results []ImageDocResult
	if allErr := cursor.All(ctx, &results); allErr != nil {
		err = errors.New("error parsing results")
		return
	}

	imgDocs = make([]dbController.ImageDocument, 0)

	for _, r := range results {
		imgDocs = append(imgDocs, r.GetImageDocument())
	}

	return
}

func (mdbc *MongoDbController) EditImageData(doc dbController.EditImageDocument) error {
	return errors.New("Unimplemented")
}

func (mdbc *MongoDbController) DeleteImage(doc dbController.DeleteImageDocument) error {
	imgDoc, imgErr := mdbc.GetImageDataById(doc.Id)

	if imgErr != nil {
		return imgErr
	}

	docId, docIdErr := primitive.ObjectIDFromHex(imgDoc.Id)
	if docIdErr != nil {
		return docIdErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	imgFileCollection := mdbc.MongoClient.Database(mdbc.dbName).Collection(IMAGE_FILE_COLLECTION)
	imgCollection := mdbc.MongoClient.Database(mdbc.dbName).Collection(IMAGE_COLLECTION)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		fileIds := make([]primitive.ObjectID, 0)

		for _, imgFile := range imgDoc.ImageFiles {
			idObj, idObjErr := primitive.ObjectIDFromHex(imgFile.Id)
			if idObjErr != nil {
				continue
			}

			fileIds = append(fileIds, idObj)
		}

		_, fErr := imgFileCollection.DeleteMany(sessCtx, bson.M{
			"_id": bson.M{
				"$in": fileIds,
			},
		})

		if fErr != nil {
			return nil, fErr
		}

		_, iErr := imgCollection.DeleteOne(sessCtx, bson.M{
			"_id": docId,
		})

		if iErr != nil {
			return nil, iErr
		}

		// db.inventory.deleteMany({
		// 	_id: {$in: [
		// 		ObjectId("61b0b36bf060f8f6ec5eda1d"),
		// 		ObjectId("61b0b36bf060f8f6ec5eda1e"),
		// 	]},
		// })

		return nil, nil
	}

	session, sessionErr := mdbc.MongoClient.StartSession()

	if sessionErr != nil {
		return sessionErr
	}
	defer session.EndSession(ctx)

	result, transErr := session.WithTransaction(ctx, callback)

	if transErr != nil {
		session.AbortTransaction(ctx)
		return transErr
	}

	fmt.Printf("result %v\n", result)

	return nil
}

func (mdbc *MongoDbController) DeleteImageFile(doc dbController.DeleteImageFileDocument) error {
	return errors.New("Unimplemented")
}

func (mdbc *MongoDbController) AddRequestLog(log logging.RequestLogData) error {
	collection, ctx, cancel := mdbc.getCollection(LOGGING_COLLECTION)
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

	_, mdbErr := collection.InsertOne(ctx, insert)

	if mdbErr != nil {
		return dbController.NewDBError(mdbErr.Error())
	}

	return nil
}

func (mdbc *MongoDbController) AddInfoLog(log logging.InfoLogData) error {
	collection, ctx, cancel := mdbc.getCollection(LOGGING_COLLECTION)
	defer cancel()

	insert := bson.M{
		"timestamp": primitive.Timestamp{T: uint32(log.Timestamp.Unix())},
		"type":      log.Type,
		"message":   log.Message,
	}

	_, mdbErr := collection.InsertOne(ctx, insert)

	if mdbErr != nil {
		return dbController.NewDBError(mdbErr.Error())
	}

	return nil
}
