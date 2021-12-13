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
	"methompson.com/image-microservice/imageServer/imageHandler"
	"methompson.com/image-microservice/imageServer/logging"
)

type MongoDbController struct {
	MongoClient *mongo.Client
	dbName      string
}

// getCollection is a convenience function that performs a function used regularly
// throughout the MongoDB controller. It accepts a collectionName string for the
// specific collection you want to retrieve, and returns a collection, context and
// cancel function.
func (mdbc *MongoDbController) getCollection(collectionName string) (*mongo.Collection, context.Context, context.CancelFunc) {
	// Write the hash to the database
	collection := mdbc.MongoClient.Database(mdbc.dbName).Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	return collection, ctx, cancel
}

// Initializes the image file collection. This collection holds information about
// image files, metadata and linking information to tie it to the image collection.
// This init function uses a schema and index to enforce the data that is supposed
// to be in individual documents.
func (mdbc *MongoDbController) initImageFileCollection(dbName string) error {
	db := mdbc.MongoClient.Database(dbName)

	// We define a schema here for image file documents
	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{
			"imageId",
			"imageIdName",
			"formatName",
			"imageType",
			"filename",
			"imageSize",
			"fileSize",
			"private",
		},
		"properties": bson.M{
			"imageId": bson.M{
				"bsonType":    "objectId",
				"description": "id of the image document to which this image belongs",
			},
			"imageIdName": bson.M{
				"bsonType":    "string",
				"description": "imageIdName is the idName of the image file",
			},
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
				"description": "imageSize must be an objet of image size data",
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

	// We set a validator for the schema defined above.
	colOpts := options.CreateCollection().SetValidator(bson.M{"$jsonSchema": jsonSchema})

	// We create the collection using the validator set above.
	createCollectionErr := db.CreateCollection(context.TODO(), IMAGE_FILE_COLLECTION, colOpts)

	if createCollectionErr != nil {
		return dbController.NewDBError(createCollectionErr.Error())
	}

	// We define an index for the image files collection. We set the index to filename
	// to ensure that all filenames are unique.
	index := []mongo.IndexModel{
		{
			Keys:    bson.M{"filename": 1},
			Options: options.Index().SetUnique(true),
		},
	}

	opts := options.CreateIndexes().SetMaxTime(2 * time.Second)
	collection, _, _ := mdbc.getCollection(IMAGE_FILE_COLLECTION)
	_, setIndexErr := collection.Indexes().CreateMany(context.TODO(), index, opts)

	if setIndexErr != nil {
		return dbController.NewDBError(setIndexErr.Error())
	}

	return nil
}

// Initializes the image collection. This collection holds information about an
// image and the various resized permutations of this image. It also contains
// metadata about the image so that we can search for it later on.
func (mdbc *MongoDbController) initImageCollection(dbName string) error {
	db := mdbc.MongoClient.Database(dbName)

	// We define a schema here for image documents
	jsonSchema := bson.M{
		"bsonType": "object",
		"required": []string{
			"title",
			"fileName",
			"idName",
			"tags",
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

	// We set a validator for the schema defined above.
	colOpts := options.CreateCollection().SetValidator(bson.M{"$jsonSchema": jsonSchema})

	// We create the collection using the validator set above.
	createCollectionErr := db.CreateCollection(context.TODO(), IMAGE_COLLECTION, colOpts)

	if createCollectionErr != nil {
		return dbController.NewDBError(createCollectionErr.Error())
	}

	// We define an index for the image files collection. We set the index to idName
	// to ensure that all filenames are unique. IdName is used to generate unique
	// file names for an image file and size variation
	index := []mongo.IndexModel{
		{
			Keys:    bson.M{"idName": 1},
			Options: options.Index().SetUnique(true),
		},
	}

	opts := options.CreateIndexes().SetMaxTime(2 * time.Second)
	collection, _, _ := mdbc.getCollection(IMAGE_COLLECTION)
	_, setIndexErr := collection.Indexes().CreateMany(context.TODO(), index, opts)

	if setIndexErr != nil {
		return dbController.NewDBError(setIndexErr.Error())
	}

	return nil
}

// Initializes the logging collection for saving logs to a database
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

// Initializes the image collection, image file collection and logging collection. It determines
// if any errors are due to the collection already existing and ignores those.
func (mdbc *MongoDbController) InitDatabase() error {
	imageCreationErr := mdbc.initImageCollection(mdbc.dbName)

	if imageCreationErr != nil && !strings.Contains(imageCreationErr.Error(), "Collection already exists") {
		return imageCreationErr
	}

	imageFileCreationErr := mdbc.initImageFileCollection(mdbc.dbName)

	if imageFileCreationErr != nil && !strings.Contains(imageFileCreationErr.Error(), "Collection already exists") {
		return imageFileCreationErr
	}

	if imageCreationErr != nil && !strings.Contains(imageCreationErr.Error(), "Collection already exists") {
		return imageCreationErr
	}

	loggingCreationErr := mdbc.initLoggingCollection(mdbc.dbName)

	if loggingCreationErr != nil && !strings.Contains(loggingCreationErr.Error(), "Collection already exists") {
		return loggingCreationErr
	}

	return nil
}

// Adds image and image file documents to the databases. The individual image files are used
// to hold information about the image files themselves, whereas the image document is used
// to hold meta information about the image and to hold the different sized versions together
// as a common set of data.
func (mdbc *MongoDbController) AddImageData(doc dbController.AddImageDocument) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	imgFileCollection := mdbc.MongoClient.Database(mdbc.dbName).Collection(IMAGE_FILE_COLLECTION)
	imgCollection := mdbc.MongoClient.Database(mdbc.dbName).Collection(IMAGE_COLLECTION)

	// The callback closure performs all of the database functions
	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		// We perform an initial check that there are images to save.
		if len(doc.SizeFormats) == 0 {
			return "", dbController.NewInvalidInputError("No images to save")
		}

		imgDoc := bson.M{
			"title":     doc.Title,
			"fileName":  doc.FileName,
			"idName":    doc.IdName,
			"tags":      doc.Tags,
			"authorId":  doc.AuthorId,
			"dateAdded": primitive.Timestamp{T: uint32(doc.DateAdded.Unix())},
		}

		// We insert a value into the image collection and check for an error
		colInsertResult, colInsertErr := imgCollection.InsertOne(ctx, imgDoc)
		if colInsertErr != nil {
			return "", dbController.NewDBError(colInsertErr.Error())
		}

		// We check the result for an error. If no error exists, we can use the image
		// ID for an input for the image files.
		imgId, idOk := colInsertResult.InsertedID.(primitive.ObjectID)
		if !idOk {
			return nil, dbController.NewDBError("invalid id returned by database")
		}

		// We build a slice of values to insert into the collection. We perform this op
		// after inserting the insert operation above in order to make sure that we have
		// a value for the imageId key.
		images := make([]interface{}, 0)
		for _, img := range doc.SizeFormats {
			var imgType string

			switch img.ImageType {
			case imageHandler.Jpeg:
				imgType = "jpeg"
			case imageHandler.Png:
				imgType = "png"
			case imageHandler.Gif:
				imgType = "gif"
			case imageHandler.Bmp:
				imgType = "bmp"
			case imageHandler.Tiff:
				imgType = "tiff"
			default:
				continue
			}

			images = append(images, bson.M{
				"imageId":     imgId,
				"imageIdName": doc.IdName,
				"formatName":  img.FormatName,
				"filename":    img.Filename,
				"imageSize": bson.M{
					"width":  img.ImageSize.Width,
					"height": img.ImageSize.Height,
				},
				"fileSize":  img.FileSize,
				"private":   img.Private,
				"imageType": imgType,
			})
		}

		// If we have no images to insert, we throw an error to rollback the writes
		// done previously. We check at this point because there's the possibility
		// that the doc.SizeFormats value has values, but each value may contain an
		// invalid input. We need to check the final slice.
		if len(images) == 0 {
			return nil, dbController.NewInvalidInputError("no images to save")
		}

		fileInsertResult, fileInsertErr := imgFileCollection.InsertMany(sessCtx, images)

		if fileInsertErr != nil {
			return nil, fileInsertErr
		}

		if len(fileInsertResult.InsertedIDs) == 0 {
			return nil, dbController.NewDBError("no images inserted")
		}

		// We return a hex string of the image document that was inserted.
		return imgId.Hex(), nil
	}

	// Here, we start a session for the purpose of making several writes at once.
	session, sessionErr := mdbc.MongoClient.StartSession()
	if sessionErr != nil {
		return "", sessionErr
	}
	defer session.EndSession(ctx)

	// We perform all actions in the session and get the results.
	result, transErr := session.WithTransaction(ctx, callback)
	if transErr != nil {
		session.AbortTransaction(ctx)
		return "", transErr
	}

	// The result should be a string representation of the id for the image document
	// inserted into the database.
	if id, ok := result.(string); ok {
		return id, nil
	} else {
		return "", errors.New("invalid database response")
	}
}

// Returns a series of bson objects that are used during the GET process for image documents,
// the image files associated with that document and user information.
func (mdbc *MongoDbController) GetImageDataAggregationStages() (projectStage, authorLookupStage, imageFileLookupStage bson.D) {
	projectStage = bson.D{
		{
			Key: "$project",
			Value: bson.M{
				"title":     1,
				"fileName":  1,
				"idName":    1,
				"tags":      1,
				"imageIds":  1,
				"authorId":  1,
				"dateAdded": 1,
				"images":    1,
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

	imageFileLookupStage = bson.D{{
		Key: "$lookup",
		Value: bson.M{
			"from":         IMAGE_FILE_COLLECTION,
			"localField":   "_id",
			"foreignField": "imageId",
			"as":           "images",
		},
	}}

	return
}

// Gets an image file from the collection with the file name provided. This should only
// produce one result, because the "filename" key is a unique index.
func (mdbc *MongoDbController) GetImageByName(name string) (imgDoc dbController.ImageFileDocument, err error) {
	collection, ctx, cancel := mdbc.getCollection(IMAGE_FILE_COLLECTION)
	defer cancel()

	var result ImageFileDocResult

	findErr := collection.FindOne(
		ctx,
		bson.M{
			"filename": name,
		},
	).Decode(&result)

	if findErr != nil {
		return imgDoc, dbController.NewDBError(findErr.Error())
	}

	return result.getImageFileDocument(), nil
}

// Gets image data (not image file data) using the id. This function constructs a matcher
// object and returns the results from GetImageDataWithMatcher
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

// GetImageDataWithMatcher is a convenience function that uses the aggregation pipeline
// to compile image documents using MongoDB's variation of an inner join.
func (mdbc *MongoDbController) GetImageDataWithMatcher(matchStage bson.D) (imgDoc dbController.ImageDocument, err error) {
	// We get the ready-made aggregation stages here that perform common search functions.
	projectStage, authorLookupStage, imageFileLookupStage := mdbc.GetImageDataAggregationStages()

	collection, ctx, cancel := mdbc.getCollection(IMAGE_COLLECTION)
	defer cancel()

	// We're expecting only one value, but we use a limit, just in case.
	// Uncertain how much this required.
	limitStage := bson.D{{
		Key:   "$limit",
		Value: int32(1),
	}}

	// The actual aggregation call. The order of the stages is important.
	cursor, aggErr := collection.Aggregate(ctx, mongo.Pipeline{
		matchStage,
		limitStage,
		authorLookupStage,
		imageFileLookupStage,
		projectStage,
	})

	if aggErr != nil {
		err = dbController.NewDBError("error getting data from database: " + aggErr.Error())
		return
	}

	// We take all of the results and place them into a slice of ImageDocResults
	var results []ImageDocResult
	if allErr := cursor.All(ctx, &results); allErr != nil {
		err = dbController.NewDBError("error parsing results: " + allErr.Error())
		return
	}

	// If there are no results, we throw the proper error
	if len(results) < 1 {
		err = dbController.NewNoResultsError("")
		return
	}

	imgDoc = results[0].GetImageDocument()
	return
}

// Retrieves information about multiple images based upon several parameters. Page and
// pagination indicate how far into the search results the collection should search.
// Pagination indicates how many results exist per page and page indicates how many results
// the program should skip before it arrives at the results it needs. sort is a struct
// that provides the function guidance on how to sort and filter the results.
func (mdbc *MongoDbController) GetImagesData(page, pagination int, sort dbController.SortImageFilter) (imgDocs []dbController.ImageDocument, err error) {
	collection, ctx, cancel := mdbc.getCollection(IMAGE_COLLECTION)
	defer cancel()

	matchStage := bson.D{{Key: "$match", Value: bson.M{}}}

	var projectStage bson.D
	projectStage, authorLookupStage, imageFileLookupStage := mdbc.GetImageDataAggregationStages()

	if !sort.ShowPrivate {
		projectStage = bson.D{
			{
				Key: "$project",
				Value: bson.M{
					"title":     1,
					"fileName":  1,
					"idName":    1,
					"tags":      1,
					"imageIds":  1,
					"authorId":  1,
					"dateAdded": 1,
					"images": bson.M{
						"$filter": bson.M{
							"input": "$images",
							"as":    "image",
							"cond": bson.M{
								"$eq": bson.A{
									"$$image.private",
									false,
								},
							},
						},
					},
				},
			},
		}
	}

	sortStage := bson.D{{
		Key:   "$sort",
		Value: bson.M{"dateAdded": -1},
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
		sortStage,
		skipStage,
		limitStage,
		authorLookupStage,
		imageFileLookupStage,
		projectStage,
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

func (mdbc *MongoDbController) GetImageFileById(id string) (doc dbController.ImageFileDocument, err error) {
	idObj, idObjErr := primitive.ObjectIDFromHex(id)

	if idObjErr != nil {
		err = dbController.NewInvalidInputError("invalid id")
		return
	}

	collection, ctx, cancel := mdbc.getCollection(IMAGE_FILE_COLLECTION)
	defer cancel()

	var result ImageFileDocResult

	findErr := collection.FindOne(
		ctx,
		bson.M{
			"_id": idObj,
		},
	).Decode(&result)

	if findErr != nil {
		return doc, findErr
	}

	return result.getImageFileDocument(), nil
}

func (mdbc *MongoDbController) ImageHasFiles(id string) (bool, error) {
	image, err := mdbc.GetImageDataById(id)

	if err != nil {
		return false, err
	}

	if len(image.ImageFiles) == 0 {
		return false, nil
	}

	return true, nil
}

func (mdbc *MongoDbController) EditImageData(doc dbController.EditImageDocument) error {
	return errors.New("Unimplemented")
}

func (mdbc *MongoDbController) EditImageFileData(doc dbController.EditImageFileDocument) (imgDoc dbController.EditImageFileResult, err error) {
	id, err := primitive.ObjectIDFromHex(doc.Id)
	if err != nil {
		return imgDoc, dbController.NewInvalidInputError("Invalid User ID")
	}

	collection, ctx, cancel := mdbc.getCollection(IMAGE_FILE_COLLECTION)
	defer cancel()

	filter := bson.M{"_id": id}

	values := bson.M{}

	if doc.ChangeObfuscate {
		values["filename"] = doc.NewName
	}

	if doc.ChangePrivate {
		values["private"] = doc.Private
	}

	update := bson.M{
		"$set": values,
	}

	_, err = collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return
	}

	return
}

func (mdbc *MongoDbController) DeleteImage(doc dbController.DeleteImageDocument) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	imgFileCollection := mdbc.MongoClient.Database(mdbc.dbName).Collection(IMAGE_FILE_COLLECTION)
	imgCollection := mdbc.MongoClient.Database(mdbc.dbName).Collection(IMAGE_COLLECTION)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		docId, docIdErr := primitive.ObjectIDFromHex(doc.Id)
		if docIdErr != nil {
			return nil, docIdErr
		}

		_, ifErr := imgFileCollection.DeleteMany(ctx, bson.M{
			"imageId": docId,
		})

		if ifErr != nil {
			return nil, ifErr
		}

		result, iErr := imgCollection.DeleteOne(ctx, bson.M{
			"_id": docId,
		})

		if iErr != nil {
			return nil, iErr
		}

		if result.DeletedCount == 0 {
			return nil, dbController.NewInvalidInputError("invalid id. no image deleted")
		}

		return nil, nil
	}

	session, sessionErr := mdbc.MongoClient.StartSession()

	if sessionErr != nil {
		return sessionErr
	}
	defer session.EndSession(ctx)

	_, transErr := session.WithTransaction(ctx, callback)

	if transErr != nil {
		session.AbortTransaction(ctx)
		return transErr
	}

	return nil
}

func (mdbc *MongoDbController) DeleteImageFile(doc dbController.DeleteImageFileDocument) error {
	imgFile, imgErr := mdbc.GetImageFileById(doc.Id)

	if imgErr != nil {
		return imgErr
	}

	docId, docIdErr := primitive.ObjectIDFromHex(doc.Id)
	if docIdErr != nil {
		return docIdErr
	}

	collection, ctx, cancel := mdbc.getCollection(IMAGE_FILE_COLLECTION)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{
		"_id": docId,
	})

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return dbController.NewInvalidInputError("invalid id. no image files deleted")
	}

	img, err := mdbc.GetImageDataById(imgFile.ImageId)

	if err != nil {
		return err
	}

	if len(img.ImageFiles) == 0 {
		did := dbController.DeleteImageDocument{Id: img.Id}
		return mdbc.DeleteImage(did)
	}

	// Put the above in a session?

	return nil
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
