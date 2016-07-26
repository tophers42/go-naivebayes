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
	conf := &Config{ModelDir: "test_files/models", Port: ":8080"}

	app = NewNaiveBayesApp(conf)

	server = httptest.NewServer(app.Handlers())
}

func unmarshalJSONResponse(t *testing.T, endpoint string, expectedStatus int, v interface{}) (response *http.Response) {
	response, responseError := http.Get(server.URL + endpoint)

	if responseError != nil {
		t.Errorf("Failed to get response from: %s. Error: %v", endpoint, responseError)
	}

	if response.StatusCode != expectedStatus {
		t.Errorf("Did not recieve expected status: %d from: %s. Recieved status: %v", expectedStatus, endpoint, response.StatusCode)
	}

	responseData, responseReadError := ioutil.ReadAll(response.Body)

	if responseReadError != nil {
		t.Errorf("Failed to read response from: %s. Error: %v", endpoint, responseReadError)
	}

	unmarshalError := json.Unmarshal(responseData, v)
	if unmarshalError != nil {
		t.Errorf("Failed to unmarshal response data from: %s. Error: %v", endpoint, unmarshalError)
	}

	return response
}

func TestCreateModel(t *testing.T) {

}

func TestViewModel(t *testing.T) {
	endpoint := "/model"

	expectedModel := &Model{}
	loadModelError := LoadFromFile(app.modelDir+"/test_model.json", expectedModel, json.Unmarshal)

	if loadModelError != nil {
		t.Errorf("Failed to load expected model from file: %v", loadModelError)
	}

	retrievedModel := &Model{}
	_ = unmarshalJSONResponse(t, endpoint+"/test_model", http.StatusOK, retrievedModel)

	if !reflect.DeepEqual(&expectedModel, &retrievedModel) {
		t.Errorf("Retrieved model (%v) did not match expected model (%v).", retrievedModel, expectedModel)
	}

	cachedModel := &Model{}
	_ = unmarshalJSONResponse(t, endpoint+"/test_model", http.StatusOK, cachedModel)

	if !reflect.DeepEqual(&retrievedModel, &cachedModel) {
		t.Errorf("Cached model (%v) did not match retrieved model (%v).", retrievedModel, expectedModel)
	}

	missingModel := &Model{}
	_ = unmarshalJSONResponse(t, endpoint+"/missing_model", http.StatusNotFound, missingModel)
}

// func TestListModels(t *testing.T) {
// 	endpoint := "/models"

// 	retrievedModels := []*Model{}
// 	unmarshalError := json.Unmarshal(responseData, &retrievedModels)
// 	if unmarshalError != nil {
// 		t.Errorf("Failed to unmarshal retrieved models: %v", unmarshalError)
// 	}

// 	expectedTestModel := &Model{}
// 	loadTestModelError := LoadFromFile(app.modelDir+"/test_model.json", expectedTestModel, json.Unmarshal)

// 	if loadTestModelError != nil {
// 		t.Errorf("Failed to load expected model from file: %v", loadTestModelError)
// 	}

// 	expectedModels := []*Model{expectedTestModel}

// 	if !reflect.DeepEqual(&expectedModels, &retrievedModels) {
// 		t.Errorf("Retrieved models (%v) did not match expected models (%v).", retrievedModels, expectedModels)
// 	}

// 	// expectedTestModel2 := &Model{}
// 	// loadTestModel2Error := LoadFromFile(app.modelDir+"/test_model2.json", expectedTestModel2, json.Unmarshal)

// 	// if loadTestModel2Error != nil {
// 	// 	t.Errorf("Failed to load expected model from file: %v", loadTestModel2Error)
// 	// }

// }
