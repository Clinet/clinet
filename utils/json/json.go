package json

import (
	//std necessities
	"encoding/json"
	"io"
)

func Marshal(source interface{}, pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(source, "", "\t")
	}
	return json.Marshal(source)
}

func Unmarshal(reader *io.Reader, target interface{}) (err error) {
	parser := json.NewDecoder(*reader)
	err = parser.Decode(&target)
	return err
}