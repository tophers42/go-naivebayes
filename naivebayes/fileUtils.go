package naivebayes

import (
	"fmt"
	"io/ioutil"
)

// SaveToFile saves the given interface to the given path using the given marshalling function
// to convert the interface into a printable []byte
func SaveToFile(path string, v interface{}, marshalFunc func(v interface{}) ([]byte, error)) (err error) {

	marshalledData, err := marshalFunc(v)
	if err != nil {
		return fmt.Errorf("Failed to save file %s: %v", path, err)
	}

	err = ioutil.WriteFile(path, marshalledData, 0755)
	if err != nil {
		return fmt.Errorf("Failed to save file %s: %v", path, err)
	}
	return nil
}

// LoadFromFile loads the data from the given path using the given unmmarshalling function
// to convert the []byte data into an interface
func LoadFromFile(path string, v interface{}, unmarshalFunc func(data []byte, v interface{}) error) (err error) {

	fileBuf, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Failed to read file %s: %v", path, err)
	}

	err = unmarshalFunc(fileBuf, v)
	if err != nil {
		return fmt.Errorf("Failed to load file %s: %v", path, err)
	}
	return nil
}
