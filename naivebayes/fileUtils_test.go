package naivebayes

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

// testStruct - struct for testing save/load operations
type testStruct struct {
	A string
	B subStruct
}

// subStruct - struct for testing nested structs
type subStruct struct {
	C string
	D int
	E map[string]int
}

// TestSaveAndLoad tests saving a struct to a file using json and
// then loading that file back into a struct
// TODO: test error cases too
func TestSaveAndLoad(t *testing.T) {
	save := testStruct{A: "abc", B: subStruct{C: "def", D: 2, E: map[string]int{"F": 5, "G": 6}}}
	saveErr := SaveToFile("test_files/test_save.json", &save, json.Marshal)
	if saveErr != nil {
		t.Errorf("Failed to save struct to file: %v", saveErr)
	}
	load := testStruct{}
	loadErr := LoadFromFile("test_files/test_save.json", &load, json.Unmarshal)
	if loadErr != nil {
		t.Errorf("Failed to load struct from file: %v", loadErr)
	}

	if !reflect.DeepEqual(&save, &load) {
		t.Errorf("Failed to save and load struct properly. Saved and loaded structs are not equal. Save: %v Load: %v", save, load)
	}
}

// TestSaveErrors tests error handling on save attempts
func TestSaveErrors(t *testing.T) {
	save := testStruct{A: "abc", B: subStruct{C: "def", D: 2}}

	failedMarshal := func(v interface{}) ([]byte, error) {
		return nil, errors.New("Marshal failed")
	}
	marshalErr := SaveToFile("test_files/valid_path.json", &save, failedMarshal)
	if marshalErr == nil {
		t.Error("Failed marshal did not throw expected error")
	}

	invalidPathErr := SaveToFile("this/is/not/a/valid/path", &save, json.Marshal)
	if invalidPathErr == nil {
		t.Error("Invalid file path did not throw expected error")
	}

}

// TestLoadErrors tests error handling on Load attempts
func TestLoadErrors(t *testing.T) {
	load := testStruct{}

	invalidPathErr := LoadFromFile("this/is/not/a/valid/path", &load, json.Unmarshal)
	if invalidPathErr == nil {
		t.Error("Invalid file path did not throw expected error")
	}

	failedUnmarshal := func(data []byte, v interface{}) error {
		return errors.New("Unmarshal failed")
	}
	marshalErr := LoadFromFile("test_files/test_save.json", &load, failedUnmarshal)
	if marshalErr == nil {
		t.Error("Failed umarshal did not throw expected error")
	}

}
