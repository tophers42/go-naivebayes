package naivebayes

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var (
	app    *NaiveBayesApp
	reader io.Reader
	server *httptest.Server
)

func init() {
	conf := &Config{TemplateDir: "templates", ModelDir: "test_files/models", Port: ":8080"}

	app = NewNaiveBayesApp(conf)

	server = httptest.NewServer(app.Handlers())
}

func TestCreateModel(t *testing.T) {

}

func TestViewExistingModel(t *testing.T) {
	response, responseError := http.Get(server.URL + "/model/test_model")

	if responseError != nil {
		t.Errorf("Failed to get response from model view. Error: %v", responseError)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Did not recieve StatusOK response from model view endpoint. Status: %v", response.StatusCode)
	}

	expectedModel := &Model{}
	loadModelError := LoadFromFile(app.modelDir+"/test_model.json", expectedModel, json.Unmarshal)

	if loadModelError != nil {
		t.Errorf("Failed to load expected model from file: %v", loadModelError)
	}

	responseData, responseReadError := ioutil.ReadAll(response.Body)

	if responseReadError != nil {
		t.Errorf("Failed to read response: %v", responseReadError)
	}

	retrievedModel := &Model{}
	unmarshalError := json.Unmarshal(responseData, retrievedModel)
	if unmarshalError != nil {
		t.Errorf("Failed to unmarshal retrieved model: %v", unmarshalError)
	}

	if !reflect.DeepEqual(&expectedModel, &retrievedModel) {
		t.Errorf("Retrieved model (%v) did not match expected model (%v).", retrievedModel, expectedModel)
	}
}

func TestViewMissingModel(t *testing.T) {
	response, responseError := http.Get(server.URL + "/model/this_model_does_not_exist")

	if responseError != nil {
		t.Errorf("Failed to get response from model view. Error: %v", responseError)
	}

	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Did not recieve StatusNotFound response from model view endpoint. Status: %v", response.StatusCode)
	}
}

func TestListModels(t *testing.T) {
	response, responseError := http.Get(server.URL + "/models")

	if responseError != nil {
		t.Errorf("Failed to get response from model view. Error: %v", responseError)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Did not recieve StatusOK response from model list endpoint. Status: %v", response.StatusCode)
	}

	responseData, responseReadError := ioutil.ReadAll(response.Body)

	if responseReadError != nil {
		t.Errorf("Failed to read response: %v", responseReadError)
	}

	retrievedModels := []*Model{}
	unmarshalError := json.Unmarshal(responseData, &retrievedModels)
	if unmarshalError != nil {
		t.Errorf("Failed to unmarshal retrieved models: %v", unmarshalError)
	}

	expectedTestModel := &Model{}
	loadTestModelError := LoadFromFile(app.modelDir+"/test_model.json", expectedTestModel, json.Unmarshal)

	if loadTestModelError != nil {
		t.Errorf("Failed to load expected model from file: %v", loadTestModelError)
	}

	expectedModels := []*Model{expectedTestModel}

	if !reflect.DeepEqual(&expectedModels, &retrievedModels) {
		t.Errorf("Retrieved models (%v) did not match expected models (%v).", retrievedModels, expectedModels)
	}

	// expectedTestModel2 := &Model{}
	// loadTestModel2Error := LoadFromFile(app.modelDir+"/test_model2.json", expectedTestModel2, json.Unmarshal)

	// if loadTestModel2Error != nil {
	// 	t.Errorf("Failed to load expected model from file: %v", loadTestModel2Error)
	// }

}
