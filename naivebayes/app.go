package naivebayes

/*
	Naivebayes classification microservice. Create and train models and run predictions
	TODO:
	* write tests for app
	* dockerize
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/gorilla/mux"
)

// JSONRequest struct
type JSONRequest struct {
	Vars        map[string]string
	QueryParams map[string][]string
	Data        []byte
}

func NewJSONRequest(r *http.Request) *JSONRequest {
	r.ParseForm()
	data, _ := ioutil.ReadAll(r.Body)
	return &JSONRequest{Vars: mux.Vars(r), Data: data, QueryParams: r.Form}
}

func (j JSONRequest) PathVar(key string) (value string) {
	return j.Vars[key]
}

func (j JSONRequest) Param(key string) (value []string) {
	return j.QueryParams[key]
}

// JSONResponse struct
type JSONResponse struct {
	Error error
	Code  int
	Data  interface{}
}

// IsError returns a boolean determining whether the response is an error.
func (j *JSONResponse) IsError() bool {
	if j.Error != nil {
		return true
	}
	return false
}

// render writes the response to the given http.ResponseWriter
func (j *JSONResponse) render(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(j.Code)
	encoder := json.NewEncoder(w)
	var err error
	if j.IsError() {
		err = encoder.Encode(j.Error)
	} else {
		err = encoder.Encode(j.Data)
	}
	if err != nil {
		log.Printf("Failed to write json response: %v.", err)
	}
}

/*
   makeHandler is a wrapper for request handling functions.
   Simple "access" logs are added to each request
   application state is provided to each handler function.
*/
func makeJSONHandler(JSONHandler func(*JSONRequest) *JSONResponse) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handling %s request to: %s", r.Method, r.URL.Path)
		jsonResponse := JSONHandler(NewJSONRequest(r))
		if jsonResponse.IsError() {
			log.Printf("%v", jsonResponse.Error)
		}
		jsonResponse.render(w)
	}
}

// Config struct
type Config struct {
	ModelDir string
	Port     string
}

// NaiveBayesApp struct
type NaiveBayesApp struct {
	modelDir string
	models   map[string]*Model
	port     string
}

// NewNaiveBayes creates and returns new App object.
// This object stores models and some configuration in memory.
func NewNaiveBayesApp(c *Config) (app *NaiveBayesApp) {

	// init a model storage dir
	os.Mkdir(c.ModelDir, 0775)

	app = &NaiveBayesApp{models: make(map[string]*Model), modelDir: c.ModelDir, port: c.Port}

	err := app.loadAllModels()
	if err != nil {
		log.Fatalf("%v", err)
	}
	return app
}

// modelPath returns the storage path for a given modelName
func (app *NaiveBayesApp) modelPath(modelName string) (path string) {
	return app.modelDir + "/" + modelName + ".json"
}

// loadAllModels loads all the models found in app.modelDir into app.models.
func (app *NaiveBayesApp) loadAllModels() (err error) {
	files, err := ioutil.ReadDir(app.modelDir)
	if err != nil {
		log.Printf("Failed to load all models from dir: '%s' with error: '%s'", app.modelDir, err)
		return err
	}
	for _, file := range files {
		model := &Model{}
		err = LoadFromFile(app.modelDir+"/"+file.Name(), model, json.Unmarshal)
		if err != nil {
			log.Printf("Failed to load model from file: '%s' with error: '%s'", file.Name(), err)
		} else {
			log.Printf("Loaded model: %s from file.", model.Name)
			app.models[model.Name] = model
		}
	}
	return nil
}

// StartServer starts the server listening on the port defined by the app object.
func (app *NaiveBayesApp) StartServer() {
	log.Printf("Listening on port: %s", app.port)
	log.Fatal(http.ListenAndServe(app.port, app.Handlers()))
}

/*
   Handlers registers all the endpoints the app will handle.
   Maps url paths to handler functions and returns a router object.
*/
func (app *NaiveBayesApp) Handlers() (router *mux.Router) {
	router = mux.NewRouter()
	router.HandleFunc("/model", makeJSONHandler(app.createModel)).Methods("POST")
	router.HandleFunc("/models", makeJSONHandler(app.listModels)).Methods("GET")
	router.HandleFunc("/model/{modelName}", makeJSONHandler(app.viewModel)).Methods("GET")
	router.HandleFunc("/model/{modelName}/train", makeJSONHandler(app.trainModel)).Methods("POST")
	router.HandleFunc("/model/{modelName}/predict", makeJSONHandler(app.predictModel)).Methods("POST")
	return router
}

/*
   ENDPOINT HANDLERS
*/

/*
   createModel creates and saves a new model from the json payload
   * POST model - Create a new model
*/
func (app *NaiveBayesApp) createModel(request *JSONRequest) *JSONResponse {
	model := &Model{}
	unmarshalErr := json.Unmarshal(request.Data, model)
	if unmarshalErr != nil {
		return &JSONResponse{Error: unmarshalErr, Code: http.StatusBadRequest}
	}

	_, exists := app.models[model.Name]

	if exists && request.Param("overwrite") == nil {
		return &JSONResponse{Error: fmt.Errorf("Could not create new model. Model %s already exists.", model.Name), Code: http.StatusConflict}
	}

	saveErr := SaveToFile(app.modelPath(model.Name), model, json.Marshal)
	if saveErr != nil {
		return &JSONResponse{Error: saveErr, Code: http.StatusInternalServerError}
	}

	app.models[model.Name] = model

	return &JSONResponse{Data: model, Code: http.StatusOK}
}

/*
   viewModel displays a model in JSON form.
   * GET model/<name> - view model
*/
func (app *NaiveBayesApp) viewModel(request *JSONRequest) *JSONResponse {
	modelName := request.PathVar("modelName")
	model, ok := app.models[modelName]

	if !ok {
		return &JSONResponse{Error: fmt.Errorf("Model not found"), Code: http.StatusNotFound}
	}

	return &JSONResponse{Data: model, Code: http.StatusOK}
}

/*
   listModels displays the list of loaded models in JSON form, in alphabetical order.
   * GET /models - Display the list of models
*/
func (app *NaiveBayesApp) listModels(request *JSONRequest) *JSONResponse {
	modelList := []*Model{}
	var sortedNames []string
	for k := range app.models {
		sortedNames = append(sortedNames, k)
	}
	sort.Strings(sortedNames)
	for _, modelName := range sortedNames {
		modelList = append(modelList, app.models[modelName])
	}
	return &JSONResponse{Data: modelList, Code: http.StatusOK}
}

/*
   trainModel displays a form for training a model with an new observation and handles
   form submission, training the given model
   * POST /model/<name/train - Trains the given model with the input observation
*/
func (app *NaiveBayesApp) trainModel(request *JSONRequest) *JSONResponse {
	modelName := request.PathVar("modelName")
	model, ok := app.models[modelName]

	if !ok {
		return &JSONResponse{Error: fmt.Errorf("Model not found"), Code: http.StatusNotFound}
	}

	observation := &Observation{}
	unmarshalErr := json.Unmarshal(request.Data, observation)
	if unmarshalErr != nil {
		return &JSONResponse{Error: unmarshalErr, Code: http.StatusBadRequest}
	}

	model.Train(observation)
	saveErr := SaveToFile(app.modelPath(model.Name), model, json.Marshal)
	if saveErr != nil {
		return &JSONResponse{Error: saveErr, Code: http.StatusInternalServerError}
	}

	log.Printf("Trained model: '%s' with new observation for classes: '%s'", model.Name, observation.Classes)
	return &JSONResponse{Data: model, Code: http.StatusOK}
}

/*
   predictModel displays a form for predicting classes
   for a new observation based on the given model
   and handles form submission returning the results of the prediction in JSON
   * POST /model/<name/predict - Predicts the class for the input using the given model
*/
func (app *NaiveBayesApp) predictModel(request *JSONRequest) *JSONResponse {
	modelName := request.PathVar("modelName")
	model, ok := app.models[modelName]

	if !ok {
		return &JSONResponse{Error: fmt.Errorf("Model not found"), Code: http.StatusNotFound}
	}

	observation := &Observation{}
	unmarshalErr := json.Unmarshal(request.Data, observation)
	if unmarshalErr != nil {
		return &JSONResponse{Error: unmarshalErr, Code: http.StatusBadRequest}
	}

	prediction := model.Predict(observation)
	return &JSONResponse{Data: prediction, Code: http.StatusOK}
}
