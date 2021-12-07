package dbController

import (
	"testing"
	"time"

	"methompson.com/image-microservice/imageServer/imageConversion"
)

func TestImageLocGetMap(t *testing.T) {
	is1 := imageConversion.ImageSize{Width: 300, Height: 200}

	ifd1 := ImageFileDocument{
		Id:         "id-1234",
		Filename:   "Test Size",
		FormatName: "image 1",
		FileSize:   128,
		ImageSize:  is1,
		Private:    true,
		ImageType:  imageConversion.Jpeg,
	}

	is2 := imageConversion.ImageSize{Width: 640, Height: 480}

	ifd2 := ImageFileDocument{
		Id:         "id-4567",
		Filename:   "Test Size Larger",
		FormatName: "image 2",
		FileSize:   256,
		ImageSize:  is2,
		Private:    false,
		ImageType:  imageConversion.Jpeg,
	}

	files := make([]ImageFileDocument, 0)
	files = append(files, ifd1)
	files = append(files, ifd2)

	imgDoc := ImageDocument{
		Id:         "123",
		Title:      "test",
		FileName:   "test.jpg",
		IdName:     "id name",
		Tags:       make([]string, 0),
		ImageFiles: files,
		Author:     "Test Author",
		AuthorId:   "456",
		DateAdded:  time.Now(),
	}

	interfaceData := imgDoc.GetMap()

	if interfaceData["id"] != "123" {
		t.Fatalf("id = '%v', Should be '123'", interfaceData["id"])
	}

	imageFiles := interfaceData["imageFiles"].([]map[string]interface{})

	if len(imageFiles) != 2 {
		t.Fatalf("len(v) = '%v', Should be '2'", len(imageFiles))
	}

	format1 := imageFiles[0]

	if format1["filename"] != "Test Size" {
		t.Fatalf("format1[\"filename\"] = '%v', Should be 'Test Size'", format1["filename"])
	}
	if format1["private"] != true {
		t.Fatalf("format1[\"private\"] = '%v', Should be 'true'", format1["private"])
	}
	if format1["fileSize"] != 128 {
		t.Fatalf("format1[\"fileSize\"] = '%v', Should be '128'", format1["fileSize"])
	}

	imageSize1 := format1["imageSize"].(map[string]interface{})

	if imageSize1["width"] != 300 {
		t.Fatalf("imageSize1[\"width\"] = '%v', Should be '300'", imageSize1["width"])
	}

	if imageSize1["height"] != 200 {
		t.Fatalf("imageSize1[\"height\"] = '%v', Should be '200'", imageSize1["height"])
	}

	imageInfo2 := imageFiles[1]

	if imageInfo2["filename"] != "Test Size Larger" {
		t.Fatalf("imageInfo2[\"filename\"] = '%v', Should be 'Test Size Larger'", imageInfo2["filename"])
	}
	if imageInfo2["private"] != false {
		t.Fatalf("imageInfo2[\"private\"] = '%v', Should be 'false'", imageInfo2["private"])
	}
	if imageInfo2["fileSize"] != 256 {
		t.Fatalf("imageInfo2[\"fileSize\"] = '%v', Should be '256'", imageInfo2["fileSize"])
	}

	imageSize2 := imageInfo2["imageSize"].(map[string]interface{})

	if imageSize2["width"] != 640 {
		t.Fatalf("imageSize2[\"width\"] = '%v', Should be '640'", imageSize2["width"])
	}

	if imageSize2["height"] != 480 {
		t.Fatalf("imageSize2[\"height\"] = '%v', Should be '480'", imageSize2["height"])
	}

}
