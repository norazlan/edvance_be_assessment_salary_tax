package pkg

import (
	"encoding/gob"
	"os"
)

// SaveToGob saves data to a GOB file
func SaveToGob(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	return encoder.Encode(data)
}

// LoadFromGob loads data from a GOB file
func LoadFromGob(filename string, data interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	return decoder.Decode(data)
}
