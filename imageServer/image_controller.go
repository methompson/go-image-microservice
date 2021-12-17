package imageServer

import (
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/gin-gonic/gin"

	"methompson.com/image-microservice/imageServer/dbController"
	"methompson.com/image-microservice/imageServer/imageHandler"
	"methompson.com/image-microservice/imageServer/logging"
)

type ImageController struct {
	DBController *dbController.DatabaseController
	Loggers      []*logging.ImageLogger
}

func InitController(dbc *dbController.DatabaseController) ImageController {
	ic := ImageController{
		DBController: dbc,
		Loggers:      make([]*logging.ImageLogger, 0),
	}

	return ic
}

func (ic *ImageController) AddLogger(logger *logging.ImageLogger) {
	ic.Loggers = append(ic.Loggers, logger)
}

func (ic *ImageController) AddImageFile(ctx *gin.Context) error {
	metaStr := ctx.PostForm("meta")
	imageFormData := parseAddImageFormString(metaStr)

	output, conversionErr := imageHandler.ProcessImageFile(ctx, imageFormData.Operations)

	if conversionErr != nil {
		return conversionErr
	}

	addImgDoc := dbController.AddImageDocument{
		Title:       imageFormData.Title,
		Tags:        imageFormData.Tags,
		IdName:      output.IdName,
		Filename:    output.OriginalFilename,
		SizeFormats: output.SizeFormats,
		AuthorId:    ctx.GetString("userId"),
		DateAdded:   time.Now(),
	}

	fmt.Println(output.OriginalFilename)
	fmt.Println(addImgDoc.AuthorId)

	_, addImageErr := (*ic.DBController).AddImageData(addImgDoc)

	if addImageErr != nil {
		// TODO Rollback database writes
		imageHandler.RollBackWrites(output)
		return addImageErr
	}

	return nil
}

func (ic *ImageController) GetImages(page, paginationNum int, sortBy string, showPrivate bool) (docs []dbController.ImageDocument, err error) {
	var _pagination int
	if paginationNum <= 0 {
		_pagination = 50
	} else {
		_pagination = paginationNum
	}

	filter := dbController.MakeSortImageFilter(sortBy)
	filter.ShowPrivate = showPrivate

	docs, err = (*ic.DBController).GetImagesData(page, _pagination, filter)
	return
}

func (ic *ImageController) GetImageByName(ctx *gin.Context) (filepath string, imgDoc dbController.ImageFileDocument, err error) {
	name := ctx.Param("imageName")

	fmt.Println(name)

	img, imgErr := (*ic.DBController).GetImageByName(name)

	if imgErr != nil {
		err = imgErr
		return
	}

	filename := img.Filename
	// Make a file path from the file name
	filepath = path.Join(imageHandler.GetImagePath(filename), filename)
	imgDoc = img

	return
}

func (ic *ImageController) GetImageDataById(id string, showPrivate bool) (doc dbController.ImageDocument, err error) {
	doc, err = (*ic.DBController).GetImageDataById(id, showPrivate)
	return
}

// When editing, we face the possibility of needing to rename the image file.
// We will branch the path off of this necessity. Both paths will eventually
// reach MakeImageFileDBEdit
func (ic *ImageController) EditImageFileDocument(doc EditImageFileBody) error {
	editDoc := doc.GetEditImageFileDocument()

	if !editDoc.ChangesExist() {
		return errors.New("no changes to make")
	}

	if editDoc.ChangeObfuscate {
		return ic.RenameImageFile(editDoc)
	}

	return ic.MakeImageFileDBEdit(editDoc)
}

// Both an file write and a DB write must be performed if the image file's
// name is being changed. We have the option of writing to the File System
// first then writing to the DB, or vice versa. Either way, if the latter
// fails, we have to rollback the former. Rolling back an image write is
// easier than rolling back a DB write, so that's the approach we take here.
func (ic *ImageController) RenameImageFile(editDoc dbController.EditImageFileDocument) error {
	// We get the image file in order to get the old file name
	imgFile, err := (*ic.DBController).GetImageFileById(editDoc.Id)

	if err != nil {
		return err
	}

	var newNameBase string

	if editDoc.Obfuscate {
		// Make a new name
		newNameBase = imageHandler.MakeRandomName()
	} else {
		// Get name from imgFileDoc
		newNameBase = imgFile.ImageIdName

	}

	// Construct the new name from the new name base above
	newName := imageHandler.MakeFilename(
		newNameBase,
		imgFile.FormatName,
		imageHandler.GetExtensionFromImageType(imgFile.ImageType),
		editDoc.Obfuscate,
	)

	// Get the new path and create the directory.
	newPath := imageHandler.GetImagePath(newName)
	err = imageHandler.CheckOrCreateImageFolder(newPath)

	if err != nil {
		return err
	}

	// Make the file pathes and move the file.
	oldFilePath := path.Join(imageHandler.GetImagePath(imgFile.Filename), imgFile.Filename)
	newFilePath := path.Join(newPath, newName)

	err = imageHandler.MoveFile(oldFilePath, newFilePath)

	if err != nil {
		return err
	}

	// Updated the edit doc with the new name and make the DB edit
	editDoc.NewName = newName

	err = ic.MakeImageFileDBEdit(editDoc)

	// If we have an error, we roll the move back
	// TODO determine how to make a compound error
	if err != nil {
		imageHandler.MoveFile(newFilePath, oldFilePath)

		return err
	}

	return nil
}

// The database component of editing an image file. We write all of the changes
// to the database.
func (ic *ImageController) MakeImageFileDBEdit(editDoc dbController.EditImageFileDocument) error {
	imgFileDoc, err := (*ic.DBController).EditImageFileData(editDoc)

	if err != nil {
		return err
	}

	fmt.Println(imgFileDoc.OldName + " -> " + imgFileDoc.NewName)

	return nil
}

func (ic *ImageController) DeleteImageDocument(delDoc dbController.DeleteImageDocument) (err error) {
	img, imgErr := (*ic.DBController).GetImageDataById(delDoc.Id, true)

	if imgErr != nil {
		return imgErr
	}

	for _, imgFile := range img.ImageFiles {
		err := DeleteFileWithImageFileDocument(imgFile)
		if err != nil {
			return err
		}
	}

	return (*ic.DBController).DeleteImage(delDoc)
}

func (ic *ImageController) DeleteImageFileDocument(delDoc dbController.DeleteImageFileDocument) (err error) {
	imgDoc, err := (*ic.DBController).DeleteImageFile(delDoc)

	if err != nil {
		return
	}

	return DeleteFileWithImageFileDocument(imgDoc)
}

func DeleteFileWithImageFileDocument(imgDoc dbController.ImageFileDocument) error {
	folderPath := imageHandler.GetImagePath(imgDoc.Filename)
	filePath := path.Join(folderPath, imgDoc.Filename)

	return imageHandler.DeleteFile(filePath)
}
