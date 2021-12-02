package imageServer

import "encoding/json"

func parseAddImageFormString(addMeta string) AddImageFormData {
	meta := GetDefaultImageFormMetaData()

	json.Unmarshal([]byte(addMeta), &meta)

	return meta
}
