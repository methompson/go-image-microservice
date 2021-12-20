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
				"description": "filename must be a string",
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
			"filename",
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
			"filename": bson.M{
				"bsonType":    "string",
				"description": "filename must be a string",
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
			"filename":  doc.Filename,
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
			return nil, dbController.NewDBError(fileInsertErr.Error())
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
		return "", dbController.NewDBError(sessionErr.Error())
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
		return "", dbController.NewDBError("invalid database response")
	}
}

// Returns a series of bson objects that are used during the GET process for image documents,
// the image files associated with that document and user information.
func (mdbc *MongoDbController) GetImageDataAggregationStages() (authorLookupStage, imageFileLookupStage bson.D) {
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

// This project stage show all values of an image, but filters out private images
func (mdbc *MongoDbController) getPublicImageProjectStage() bson.D {
	return bson.D{
		{
			Key: "$project",
			Value: bson.M{
				"title":     1,
				"filename":  1,
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

// This projection stage shows all values of an image and makes no distinction
// between public and private images.
func (mdbc *MongoDbController) getImageProjectStage() bson.D {
	return bson.D{
		{
			Key: "$project",
			Value: bson.M{
				"title":     1,
				"filename":  1,
				"idName":    1,
				"tags":      1,
				"imageIds":  1,
				"authorId":  1,
				"dateAdded": 1,
				"images":    1,
			},
		},
	}
}

// Gets an image file from the collection with the file name provided. This should only
// produce one result, because the "filename" key is a unique index.
func (mdbc *MongoDbController) GetImageByName(name string) (imgDoc dbController.ImageFileDocument, err error) {
	collection, ctx, cancel := mdbc.getCollection(IMAGE_FILE_COLLECTION)
	defer cancel()

	var result ImageFileDocResult

	err = collection.FindOne(
		ctx,
		bson.M{
			"filename": name,
		},
	).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return imgDoc, dbController.NewNoResultsError("")
		}
		return imgDoc, dbController.NewDBError(err.Error())
	}

	return result.getImageFileDocument(), nil
}

// Gets image data (not image file data) using the id. This function constructs a matcher
// object and returns the results from GetImageDataWithMatcher
func (mdbc *MongoDbController) GetImageDataById(id string, showPrivate bool) (imgDoc dbController.ImageDocument, err error) {
	idObj, idObjErr := primitive.ObjectIDFromHex(id)

	if idObjErr != nil {
		err = dbController.NewInvalidInputError("invalid id")
		return
	}

	matchStage := mdbc.getBasicMatcher(idObj)

	collection, ctx, cancel := mdbc.getCollection(IMAGE_COLLECTION)
	defer cancel()

	return mdbc.getImageDataWithMatcherAndCollection(matchStage, showPrivate, collection, ctx)
}

// This function mirrors the above function, but also accepts a context for use
// with session transactions
func (mdbc *MongoDbController) getImageDataByIdWithCollection(id string, collection *mongo.Collection, ctx context.Context) (imgDoc dbController.ImageDocument, err error) {
	idObj, idObjErr := primitive.ObjectIDFromHex(id)

	if idObjErr != nil {
		err = dbController.NewInvalidInputError("invalid id")
		return
	}

	matchStage := mdbc.getBasicMatcher(idObj)

	return mdbc.getImageDataWithMatcherAndCollection(matchStage, true, collection, ctx)
}

// The basic matcher is used a couple times. By moving assignment of the matcher to its
// own function we reduce some of the 'noise' in the application.
func (mdbc *MongoDbController) getBasicMatcher(idObj primitive.ObjectID) bson.D {
	return bson.D{
		{
			Key: "$match",
			Value: bson.M{
				"_id": idObj,
			},
		},
	}
}

// GetImageDataWithMatcher is a convenience function that uses the aggregation pipeline
// to compile image documents using MongoDB's variation of an inner join.
func (mdbc *MongoDbController) getImageDataWithMatcherAndCollection(matchStage bson.D, showPrivate bool, collection *mongo.Collection, ctx context.Context) (imgDoc dbController.ImageDocument, err error) {
	// We're expecting only one value, but we use a limit, just in case.
	// Uncertain how much this required.
	limitStage := bson.D{{
		Key:   "$limit",
		Value: int32(1),
	}}

	// We get the ready-made aggregation stages here that perform common search functions.
	authorLookupStage, imageFileLookupStage := mdbc.GetImageDataAggregationStages()

	// We use the showPrivate boolean to determine whether we should use the more
	// or less permissive projection stage
	var projectStage bson.D
	if showPrivate {
		projectStage = mdbc.getImageProjectStage()
	} else {
		projectStage = mdbc.getPublicImageProjectStage()
	}

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

	// The aggregation pipeline. Essentially a mutable slice
	pipeline := mongo.Pipeline{}

	// The match stage is for matching by nothing in particular
	pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.M{}}})

	// This is the sort stage
	switch sort.Sortby {
	case dbController.Name:
		pipeline = append(pipeline, mdbc.getLowerCaseStage())
		pipeline = append(pipeline, bson.D{{
			Key:   "$sort",
			Value: bson.M{"lowercaseFilename": 1},
		}})
	case dbController.NameReverse:
		pipeline = append(pipeline, mdbc.getLowerCaseStage())
		pipeline = append(pipeline, bson.D{{
			Key:   "$sort",
			Value: bson.M{"lowercaseFilename": -1},
		}})
	case dbController.DateAddedReverse:
		pipeline = append(pipeline, bson.D{{
			Key:   "$sort",
			Value: bson.M{"dateAdded": 1},
		}})
	// case dbController.DateAdded:
	default:
		pipeline = append(pipeline, bson.D{{
			Key:   "$sort",
			Value: bson.M{"dateAdded": -1},
		}})
	}

	// This is the skip stage. We skip based upon pagination and current page.
	pipeline = append(pipeline, bson.D{{
		Key:   "$skip",
		Value: int64((page - 1) * pagination),
	}})

	// This is the limit stage. We limit based upon pagination (how many results per page)
	pipeline = append(pipeline, bson.D{{
		Key:   "$limit",
		Value: int32(pagination),
	}})

	// We get the common author and image file lookup stages
	authorLookupStage, imageFileLookupStage := mdbc.GetImageDataAggregationStages()

	pipeline = append(pipeline, authorLookupStage)
	pipeline = append(pipeline, imageFileLookupStage)

	// This is the projection stage. We use the showPrivate boolean to determine whether
	// we should use the more or less permissive projection stage
	if !sort.ShowPrivate {
		pipeline = append(pipeline, mdbc.getPublicImageProjectStage())
	} else {
		pipeline = append(pipeline, mdbc.getImageProjectStage())

	}

	// The aggregation stages:
	// matchStage
	// lowerCaseLettersStage
	// sortStage
	// skipStage
	// limitStage
	// authorLookupStage
	// imageFileLookupStage
	// projectStage
	cursor, err := collection.Aggregate(ctx, pipeline)

	if err != nil {
		err = dbController.NewDBError("error getting results")
		return
	}

	var results []ImageDocResult
	if err = cursor.All(ctx, &results); err != nil {
		err = dbController.NewDBError("error parsing results")
		return
	}

	imgDocs = make([]dbController.ImageDocument, 0)

	for _, r := range results {
		imgDocs = append(imgDocs, r.GetImageDocument())
	}

	return
}

// This stage is used to generate lower case letters for each file name for when we're
// sorting by file name.
func (mdbc *MongoDbController) getLowerCaseStage() bson.D {
	return bson.D{{
		Key: "$project",
		Value: bson.M{
			"filename": 1,
			"lowercaseFilename": bson.M{
				"$toLower": "$filename",
			},
		},
	}}
}

// Gets an image file by its ID.
func (mdbc *MongoDbController) GetImageFileById(id string) (imgDoc dbController.ImageFileDocument, err error) {
	collection, ctx, cancel := mdbc.getCollection(IMAGE_FILE_COLLECTION)
	defer cancel()

	idObj, idObjErr := primitive.ObjectIDFromHex(id)

	if idObjErr != nil {
		err = dbController.NewInvalidInputError("invalid id")
		return
	}

	var result ImageFileDocResult

	err = collection.FindOne(
		ctx,
		bson.M{
			"_id": idObj,
		},
	).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return imgDoc, dbController.NewNoResultsError("")
		}
		return imgDoc, dbController.NewDBError(err.Error())
	}

	return result.getImageFileDocument(), nil
}

// This function determines if an image has any files. This is to be used for the
// cleanup function.
func (mdbc *MongoDbController) ImageHasFiles(id string) (bool, error) {
	image, err := mdbc.GetImageDataById(id, true)

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

// This function edits an individual image file. Currently, there are two factors
// that can be changed: We can make the image private or not private and we can
// obfuscate the file name (to prevent people from guessing the image files)
func (mdbc *MongoDbController) EditImageFileData(doc dbController.EditImageFileDocument) (imgDoc dbController.EditImageFileResult, err error) {
	// If there are no edits to be made, we'll return an error
	if !doc.ChangesExist() {
		return imgDoc, dbController.NewInvalidInputError("no edits to be made")
	}

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

// This function deletes an image document, including the files associated with it
func (mdbc *MongoDbController) DeleteImage(doc dbController.DeleteImageDocument) error {
	docId, docIdErr := primitive.ObjectIDFromHex(doc.Id)
	if docIdErr != nil {
		return docIdErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	imgFileCollection := mdbc.MongoClient.Database(mdbc.dbName).Collection(IMAGE_FILE_COLLECTION)
	imgCollection := mdbc.MongoClient.Database(mdbc.dbName).Collection(IMAGE_COLLECTION)

	// We use a session to make the writes.
	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		_, ifErr := imgFileCollection.DeleteMany(ctx, bson.M{
			"imageId": docId,
		})

		if ifErr != nil {
			return nil, ifErr
		}

		// result, iErr := imgCollection.DeleteOne(ctx, bson.M{
		// 	"_id": docId,
		// })

		// if iErr != nil {
		// 	return nil, iErr
		// }

		// if result.DeletedCount == 0 {
		// 	return nil, dbController.NewInvalidInputError("invalid id. no image deleted")
		// }

		err := mdbc.deleteImageDataWithContext(docId, imgCollection, sessCtx)

		if err != nil {
			return nil, err
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

// This is a convenience function that performs the delete operation given the the
// values required to perform the delete function. Some functions want to run this
// set of operations without the set up found in DeleteImage.
func (mdbc *MongoDbController) deleteImageDataWithContext(docId primitive.ObjectID, imgCollection *mongo.Collection, ctx context.Context) (err error) {
	result, err := imgCollection.DeleteOne(ctx, bson.M{
		"_id": docId,
	})

	if err != nil {
		return
	}

	if result.DeletedCount == 0 {
		return dbController.NewInvalidInputError("invalid id. no image deleted")
	}

	return
}

// This deletes an image file document. If it deletes the last image associated
// with an image document, it also deletes the image document.
func (mdbc *MongoDbController) DeleteImageFile(doc dbController.DeleteImageFileDocument) (imgDoc dbController.ImageFileDocument, err error) {
	docId, err := primitive.ObjectIDFromHex(doc.Id)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	imgFileCollection := mdbc.MongoClient.Database(mdbc.dbName).Collection(IMAGE_FILE_COLLECTION)
	imgCollection := mdbc.MongoClient.Database(mdbc.dbName).Collection(IMAGE_COLLECTION)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		var result ImageFileDocResult

		// We use the FindOneAndDelete function to get the value we just deleted so
		// that we can do three things with it:
		// Get the ImageId and determine if any more images exist for that ImageId
		// Delete the Image if no images exist for that ImageId
		// Get the filename so we can delete the Image file
		err = imgFileCollection.FindOneAndDelete(sessCtx, bson.M{
			"_id": docId,
		}).Decode(&result)

		if err != nil {
			return nil, err
		}

		// imgDoc is the eventual returned value
		imgDoc = result.getImageFileDocument()

		img, err := mdbc.getImageDataByIdWithCollection(imgDoc.ImageId, imgCollection, sessCtx)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, dbController.NewNoResultsError("")
			}
			return nil, err
		}

		// We run this if statement if the image files left for this image are zero.
		if len(img.ImageFiles) == 0 {
			// Here, we get an ObjectID from the hex value, and call the delete function
			// get rid of the image data.
			docId, docIdErr := primitive.ObjectIDFromHex(img.Id)
			if docIdErr != nil {
				return nil, docIdErr
			}

			return imgDoc, mdbc.deleteImageDataWithContext(docId, imgCollection, sessCtx)
		}

		return imgDoc, nil
	}

	// Start the session, defer ending the session, run the transaction in the session
	session, err := mdbc.MongoClient.StartSession()

	if err != nil {
		return
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, callback)

	if err != nil {
		// I'm not sure this function is required
		session.AbortTransaction(ctx)
		return
	}

	return
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
