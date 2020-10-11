package utils

import (
	"encoding/json"
	"io/ioutil"
)

// WriteJSONFile will write a struct out as JSON to the specified file
func WriteJSONFile(data interface{}, filename string) error {
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, file, 0777)
}

func WriteFile(data string, filename string) error {
	return ioutil.WriteFile(filename, []byte(data), 0777)
}
