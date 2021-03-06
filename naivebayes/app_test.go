package naivebayes

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"
)

var (
	app         *NaiveBayesApp
	reader      io.Reader
	server      *httptest.Server
	invalidJSON []byte
)

func init() {
	conf := &Config{ModelDir: "test_files/models", Port: ":8080"}

	app = NewNaiveBayesApp(conf)

	server = httptest.NewServer(app.Handlers())

	invalidJSON = []byte("{]{]{]this is not valid json![}[}[}")
}

func cleanupModel(t *testing.T, modelName string) {
	delete(app.models, modelName)
	cleanUpErr := os.Remove(app.modelPath(modelName))
	if cleanUpErr != nil {
		t.Fatalf("Failed to clean up model: %s test: %v", modelName, cleanUpErr)
	}
}

func unmarshalJSONResponse(t *testing.T, request *http.Request, expectedStatus int, v interface{}) (response *http.Response) {

	response, responseErr := http.DefaultClient.Do(request)

	if responseErr != nil {
		t.Errorf("Failed to get response from: %v. Error: %v", request.URL, responseErr)
	}

	responseData, responseReadErr := ioutil.ReadAll(response.Body)

	if responseReadErr != nil {
		t.Errorf("Failed to read response from: %v. Error: %v", request.URL, responseReadErr)
	}

	if response.StatusCode != expectedStatus {
		t.Errorf("Did not recieve expected status: %d from request: %v. Recieved status: %v", expectedStatus, request.URL, response.StatusCode)
	}

	unmarshalErr := json.Unmarshal(responseData, v)
	if unmarshalErr != nil {
		t.Errorf("Failed to unmarshal response data from request: %v. Error: %v", request.URL, unmarshalErr)
	}

	return response
}

func TestCreateModel(t *testing.T) {
	endpoint := server.URL + "/model"
	expectedModel := NewModel("create_model")
	observation := NewObservationFromText([]string{"test"}, "extra data")
	createModelJSON, _ := json.Marshal(expectedModel)

	createRequest, createRequestErr := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(createModelJSON))
	if createRequestErr != nil {
		t.Errorf("Failed to generate request: %v", createRequestErr)
	}

	createdModel := &Model{}
	_ = unmarshalJSONResponse(t, createRequest, http.StatusOK, createdModel)

	if !reflect.DeepEqual(&expectedModel, &createdModel) {
		t.Errorf("Retrieved model (%v) did not match expected model (%v).", createdModel, expectedModel)
	}

	recreateRequest, recreateRequestErr := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(createModelJSON))
	if recreateRequestErr != nil {
		t.Errorf("Failed to generate request: %v", recreateRequestErr)
	}
	_ = unmarshalJSONResponse(t, recreateRequest, http.StatusConflict, createdModel)

	overwriteModel := NewModel("create_model")
	overwriteModel.Train(observation)
	overwriteModelJSON, _ := json.Marshal(overwriteModel)
	overwriteParam := url.Values{}
	overwriteParam.Set("overwrite", "1")
	overwriteRequest, overwriteRequestErr := http.NewRequest(http.MethodPost, endpoint+"?"+overwriteParam.Encode(), bytes.NewBuffer(overwriteModelJSON))
	if overwriteRequestErr != nil {
		t.Errorf("Failed to generate request: %v", overwriteRequestErr)
	}
	overwrittenModel := &Model{}
	_ = unmarshalJSONResponse(t, overwriteRequest, http.StatusOK, overwrittenModel)
	if !reflect.DeepEqual(&overwriteModel, &overwrittenModel) {
		t.Errorf("Overwritten model (%v) did not match expected model (%v).", overwriteModel, overwrittenModel)
	}

	invalidModel := &Model{}
	invalidRequest, invalidRequestErr := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(invalidJSON))
	if invalidRequestErr != nil {
		t.Errorf("Failed to generate request: %v", invalidRequestErr)
	}
	_ = unmarshalJSONResponse(t, invalidRequest, http.StatusBadRequest, invalidModel)

	cleanupModel(t, "create_model")
}

func TestViewModel(t *testing.T) {
	endpoint := server.URL + "/model"

	expectedModel := &Model{}
	loadModelErr := LoadFromFile(app.modelDir+"/test_model.json", expectedModel, json.Unmarshal)

	if loadModelErr != nil {
		t.Errorf("Failed to load expected model from file: %v", loadModelErr)
	}

	retrievedModel := &Model{}
	retrieveRequest, retrieveRequestErr := http.NewRequest(http.MethodGet, endpoint+"/test_model", nil)
	if retrieveRequestErr != nil {
		t.Errorf("Failed to generate request: %v", retrieveRequestErr)
	}
	_ = unmarshalJSONResponse(t, retrieveRequest, http.StatusOK, retrievedModel)

	if !reflect.DeepEqual(&expectedModel, &retrievedModel) {
		t.Errorf("Retrieved model (%v) did not match expected model (%v).", retrievedModel, expectedModel)
	}

	missingModel := &Model{}
	missingRequest, missingRequestErr := http.NewRequest(http.MethodGet, endpoint+"/missing_model", nil)
	if missingRequestErr != nil {
		t.Errorf("Failed to generate request: %v", missingRequestErr)
	}
	_ = unmarshalJSONResponse(t, missingRequest, http.StatusNotFound, missingModel)
}

func TestListModels(t *testing.T) {
	endpoint := server.URL + "/models"

	testModel := &Model{}
	loadTestModelErr := LoadFromFile(app.modelDir+"/test_model.json", testModel, json.Unmarshal)
	if loadTestModelErr != nil {
		t.Errorf("Failed to load test model from file: %v", loadTestModelErr)
	}
	emptyModel := &Model{}
	loadEmptyModelErr := LoadFromFile(app.modelDir+"/empty_model.json", emptyModel, json.Unmarshal)
	if loadEmptyModelErr != nil {
		t.Errorf("Failed to load test model from file: %v", loadEmptyModelErr)
	}
	expectedModels := []*Model{emptyModel, testModel}

	retrievedModels := []*Model{}
	retrieveRequest, retrieveRequestErr := http.NewRequest(http.MethodGet, endpoint, nil)
	if retrieveRequestErr != nil {
		t.Errorf("Failed to generate request: %v", retrieveRequestErr)
	}
	_ = unmarshalJSONResponse(t, retrieveRequest, http.StatusOK, &retrievedModels)

	if !reflect.DeepEqual(&expectedModels, &retrievedModels) {
		t.Errorf("Retrieved models (%v) did not match expected models (%v).", retrievedModels, expectedModels)
	}
}

func TestTrainModel(t *testing.T) {
	// setup
	trainModel := NewModel("train_model")
	trainModelJSON, _ := json.Marshal(trainModel)
	createRequest, createRequestErr := http.NewRequest(http.MethodPost, server.URL+"/model", bytes.NewBuffer(trainModelJSON))
	if createRequestErr != nil {
		t.Errorf("Failed to generate request: %v", createRequestErr)
	}
	http.DefaultClient.Do(createRequest)

	testObservation := NewObservationFromText([]string{"testing"}, "test observation")
	testObservationJSON, _ := json.Marshal(testObservation)
	trainRequest, trainRequestErr := http.NewRequest(http.MethodPost, server.URL+"/model/train_model/train", bytes.NewBuffer(testObservationJSON))
	if trainRequestErr != nil {
		t.Errorf("Failed to generate request: %v", trainRequestErr)
	}
	trainedModel := &Model{}
	_ = unmarshalJSONResponse(t, trainRequest, http.StatusOK, &trainedModel)

	trainModel.Train(testObservation)
	if !reflect.DeepEqual(&trainModel, &trainedModel) {
		t.Errorf("Retrieved trained (%v) did not match expected model (%v).", trainedModel, trainModel)
	}

	missingRequest, missingRequestErr := http.NewRequest(http.MethodPost, server.URL+"/model/missing_model/train", bytes.NewBuffer(testObservationJSON))
	if missingRequestErr != nil {
		t.Errorf("Failed to generate request: %v", missingRequestErr)
	}
	missingModel := &Model{}
	_ = unmarshalJSONResponse(t, missingRequest, http.StatusNotFound, &missingModel)

	invalidRequest, invalidRequestErr := http.NewRequest(http.MethodPost, server.URL+"/model/train_model/train", bytes.NewBuffer(invalidJSON))
	if invalidRequestErr != nil {
		t.Errorf("Failed to generate request: %v", invalidRequestErr)
	}
	invalidModel := &Model{}
	_ = unmarshalJSONResponse(t, invalidRequest, http.StatusBadRequest, &invalidModel)

	cleanupModel(t, "train_model")
}
