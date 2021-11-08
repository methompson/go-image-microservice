package dbController

import (
	"testing"
	"time"
)

func TestImageLocGetMap(t *testing.T) {
	is1 := ImageSize{Width: 300, Height: 200}

	il1 := ImageLocation{
		SizeType:  "Test Size",
		Url:       "/path/to/test",
		FileSize:  128,
		ImageSize: &is1,
	}

	is2 := ImageSize{Width: 640, Height: 480}

	il2 := ImageLocation{
		SizeType:  "Test Size Larger",
		Url:       "/path/to/test2",
		FileSize:  256,
		ImageSize: &is2,
	}

	locations := make([]*ImageLocation, 0)
	locations = append(locations, &il1)
	locations = append(locations, &il2)

	id := ImageDocument{
		Id:             "123",
		Title:          "test",
		FileName:       "test.jpg",
		Locations:      locations,
		Author:         "Test Author",
		AuthorId:       "456",
		DateAdded:      time.Now(),
		UpdateAuthor:   "Test Update Author",
		UpdateAuthorId: "789",
		DateUpdated:    time.UnixMilli(0),
	}

	interfaceData := *id.GetMap()

	if interfaceData["id"] != "123" {
		t.Fatalf("id = '%v', Should be '123'", interfaceData["id"])
	}

	v := interfaceData["locations"].([]*map[string]interface{})

	if len(v) != 2 {
		t.Fatalf("len(v) = '%v', Should be '2'", len(v))
	}

	imageInfo1 := *v[0]

	if imageInfo1["sizeType"] != "Test Size" {
		t.Fatalf("imageInfo1[\"sizeType\"] = '%v', Should be 'Test Size'", imageInfo1["sizeType"])
	}
	if imageInfo1["url"] != "/path/to/test" {
		t.Fatalf("imageInfo1[\"url\"] = '%v', Should be '/path/to/test'", imageInfo1["url"])
	}
	if imageInfo1["fileSize"] != 128 {
		t.Fatalf("imageInfo1[\"fileSize\"] = '%v', Should be '128'", imageInfo1["v"])
	}

	imageSize1 := *(imageInfo1["imageSize"].(*map[string]interface{}))

	if imageSize1["width"] != 300 {
		t.Fatalf("imageSize1[\"width\"] = '%v', Should be '300'", imageSize1["width"])
	}

	if imageSize1["height"] != 200 {
		t.Fatalf("imageSize1[\"height\"] = '%v', Should be '200'", imageSize1["height"])
	}

	imageInfo2 := *v[1]

	if imageInfo2["sizeType"] != "Test Size Larger" {
		t.Fatalf("imageInfo2[\"sizeType\"] = '%v', Should be 'Test Size Larger'", imageInfo2["sizeType"])
	}
	if imageInfo2["url"] != "/path/to/test2" {
		t.Fatalf("imageInfo2[\"url\"] = '%v', Should be '/path/to/test2'", imageInfo2["url"])
	}
	if imageInfo2["fileSize"] != 256 {
		t.Fatalf("imageInfo2[\"fileSize\"] = '%v', Should be '256'", imageInfo2["v"])
	}

	imageSize2 := *(imageInfo2["imageSize"].(*map[string]interface{}))

	if imageSize2["width"] != 640 {
		t.Fatalf("imageSize2[\"width\"] = '%v', Should be '640'", imageSize2["width"])
	}

	if imageSize2["height"] != 480 {
		t.Fatalf("imageSize2[\"height\"] = '%v', Should be '480'", imageSize2["height"])
	}

}
