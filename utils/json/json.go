package json

import (
	//std necessities
	"encoding/json"
)

func Marshal(source interface{}, pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(source, "", "\t")
	}
	return json.Marshal(source)
}

func Unmarshal(dataJSON []byte, target interface{}) (err error) {
	err = json.Unmarshal(dataJSON, target)
	return
}