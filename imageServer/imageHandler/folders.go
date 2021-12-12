package imageHandler

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"strconv"

	"methompson.com/image-microservice/imageServer/constants"
)

// Gets the image path from the env variables. If the image path does not
// exist as an env variable, we return a default location.
func GetImageRootPath() string {
	envPath := os.Getenv(constants.IMAGE_PATH)

	if len(envPath) > 0 {
		return envPath
	}

	return "./files"
}

// We won't assume that the filename is always UUID, so we are going to do the following:
// * Check for a hyphen and get all text prior to the hyphen
// * If a hyphen doesn't exist, get the characters prior to the first dot
// * Choose the first X characters of that name
// * Create a folder with those X characters
// * return the path to that folder
// X is defined by an environment variable, but defaults to 2
func GetImagePath(filename string) string {
	var subfolder string

	subPathLength := getSubPathLength()
	if len(filename) <= subPathLength {
		subfolder = filename
	} else {
		subfolder = filename[:subPathLength]
	}

	folderPath := path.Join(GetImageRootPath(), subfolder)

	return folderPath
}

func getSubPathLength() int {
	val, err := strconv.Atoi(os.Getenv(constants.IMAGE_SUB_PATH_LENGTH))

	if err != nil || val < 1 {
		return 2
	}

	return val
}

// Checks if the image file location exists. If it does not
// the application will create the folder. This should be a
// one time thing
func CheckOrCreateImageFolder(folderPath string) error {
	stat, statErr := os.Stat(folderPath)
	if os.IsNotExist(statErr) {
		// Image path does not exist
		mkdirErr := os.MkdirAll(folderPath, 0740)

		if mkdirErr != nil {
			return mkdirErr
		}

		return nil
	}

	if isReadable(stat.Mode()) && isWriteable(stat.Mode()) {
		return nil
	}
	return errors.New("cannot read and/or write to " + folderPath)
}

// Checks if the mode allows for the owner to read
func isReadable(mode fs.FileMode) bool {
	// 0200
	return mode&OS_USER_W != 0
}

// Checks if the mode allows for anyone to read
func isWriteable(mode fs.FileMode) bool {
	// 0444
	return mode&OS_ALL_R != 0
}

const (
	OS_READ        = 04
	OS_WRITE       = 02
	OS_EX          = 01
	OS_USER_SHIFT  = 6
	OS_GROUP_SHIFT = 3
	OS_OTH_SHIFT   = 0

	OS_USER_R   = OS_READ << OS_USER_SHIFT
	OS_USER_W   = OS_WRITE << OS_USER_SHIFT
	OS_USER_X   = OS_EX << OS_USER_SHIFT
	OS_USER_RW  = OS_USER_R | OS_USER_W
	OS_USER_RX  = OS_USER_R | OS_USER_X
	OS_USER_WX  = OS_USER_W | OS_USER_X
	OS_USER_RWX = OS_USER_RW | OS_USER_X

	OS_GROUP_R   = OS_READ << OS_GROUP_SHIFT
	OS_GROUP_W   = OS_WRITE << OS_GROUP_SHIFT
	OS_GROUP_X   = OS_EX << OS_GROUP_SHIFT
	OS_GROUP_RW  = OS_GROUP_R | OS_GROUP_W
	OS_GROUP_RX  = OS_GROUP_R | OS_GROUP_X
	OS_GROUP_WX  = OS_GROUP_W | OS_GROUP_X
	OS_GROUP_RWX = OS_GROUP_RW | OS_GROUP_X

	OS_OTH_R   = OS_READ << OS_OTH_SHIFT
	OS_OTH_W   = OS_WRITE << OS_OTH_SHIFT
	OS_OTH_X   = OS_EX << OS_OTH_SHIFT
	OS_OTH_RW  = OS_OTH_R | OS_OTH_W
	OS_OTH_RX  = OS_OTH_R | OS_OTH_X
	OS_OTH_WX  = OS_OTH_W | OS_OTH_X
	OS_OTH_RWX = OS_OTH_RW | OS_OTH_X

	OS_ALL_R   = OS_USER_R | OS_GROUP_R | OS_OTH_R
	OS_ALL_W   = OS_USER_W | OS_GROUP_W | OS_OTH_W
	OS_ALL_X   = OS_USER_X | OS_GROUP_X | OS_OTH_X
	OS_ALL_RW  = OS_ALL_R | OS_ALL_W
	OS_ALL_RWX = OS_ALL_RW | OS_GROUP_X
)
