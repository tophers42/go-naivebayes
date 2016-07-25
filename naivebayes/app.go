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
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

// Config struct
type Config struct {
	TemplateDir string
	ModelDir    string
	Port        string
}

// NaiveBayesApp struct
type NaiveBayesApp struct {
	templates *template.Template
	modelDir  string
	models    map[string]*Model
	port      string
}

// NewNaiveBayes creates and returns new App object.
// This object stores models and some configuration in memory.
func NewNaiveBayesApp(c *Config) (app *NaiveBayesApp) {

	// init a model storage dir
	os.Mkdir(c.ModelDir, 0775)

	// init the templates
	var templates = template.Must(template.ParseGlob(c.TemplateDir + "/*"))

	return &NaiveBayesApp{templates: templates, models: make(map[string]*Model), modelDir: c.ModelDir, port: c.Port}
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
	router.HandleFunc("/model", app.makeHandler(app.createModel)).Methods("GET", "POST")
	router.HandleFunc("/models", app.makeHandler(app.listModels)).Methods("GET")
	router.HandleFunc("/model/{modelName}", app.makeHandler(app.viewModel)).Methods("GET")
	router.HandleFunc("/model/{modelName}/train", app.makeHandler(app.trainModel)).Methods("GET", "POST")
	router.HandleFunc("/model/{modelName}/predict", app.makeHandler(app.predictModel)).Methods("GET", "POST")
	return router
}

/*
   makeHandler is a wrapper for request handling functions.
   Simple "access" logs are added to each request
   application state is provided to each handler function.
*/
func (app *NaiveBayesApp) makeHandler(handlerFunc func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		log.Printf("Handling request to: %s", request.URL.Path)
		handlerFunc(response, request)
	}
}

// renderTemplate renders the given template,
// throwing an http error if there's an issue.
func (app *NaiveBayesApp) renderTemplate(response http.ResponseWriter, tmpl string, data interface{}) {
	err := app.templates.ExecuteTemplate(response, tmpl, data)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

// modelPath returns the storage path for a given modelName
func (app *NaiveBayesApp) modelPath(modelName string) (path string) {
	return app.modelDir + "/" + modelName + ".json"
}

/*
   loadModel loads and returns the model instance for the given modelName.
   First checks app.models for an in memory copy, then tries to load the model from a file
*/
func (app *NaiveBayesApp) loadModel(modelName string) (model *Model, err error) {
	model, ok := app.models[modelName]
	if ok {
		return model, nil
	}
	log.Printf("Model: '%s' not found in memory. Falling back to file.", modelName)
	model = &Model{}
	err = LoadFromFile(app.modelPath(modelName), model, json.Unmarshal)
	if err == nil {
		app.models[modelName] = model
		return model, nil
	}
	log.Printf("Failed to load model from file: '%s', err: '%s'", modelName, err)
	return nil, fmt.Errorf("Unable to load model: '%s'", modelName)
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

// clearLoadedModels clears all the loaded models from app.models.
func (app *NaiveBayesApp) clearCachedModels() {
	app.models = map[string]*Model{}
}

/*
   ENDPOINT HANDLERS
*/

/*
   createModel displays a model creation form and handles form submission,
   creating a new model based on input
   * GET model - Display new model form
   * POST model - Create a new model
*/
func (app *NaiveBayesApp) createModel(response http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		app.renderTemplate(response, "model_create.html", nil)
	} else if request.Method == "POST" {
		request.ParseForm()
		modelName := request.FormValue("model_name")
		modelExists, _ := app.loadModel(modelName)
		if modelExists != nil {
			fmt.Fprintf(response, "Model by that name already exists: '%s'", modelName)
		} else {
			model := NewModel(modelName)
			err := SaveToFile("saved_models/"+modelName+".json", model, json.Marshal)
			if err != nil {
				http.Error(response, err.Error(), http.StatusInternalServerError)
			} else {
				log.Printf("Created new model: '%s'", modelName)
				app.models[modelName] = model
				http.Redirect(response, request, "/model/"+modelName, 301)
			}
		}
	}
}

/*
   viewModel displays a model in JSON form.
   * GET model/<name> - view model
*/
func (app *NaiveBayesApp) viewModel(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	modelName := vars["modelName"]
	model, err := app.loadModel(modelName)

	if err == nil {
		jsonModel, _ := json.Marshal(model)
		fmt.Fprintf(response, string(jsonModel))
	} else {
		http.Error(response, fmt.Sprintf("Unable to load model: '%s'", modelName), http.StatusNotFound)
	}
}

/*
   listModels displays the list of loaded models in JSON form.
   Handles "load_all" query param to force loading of all models from files into memory.
   * GET /models - Display the list of models
*/
func (app *NaiveBayesApp) listModels(response http.ResponseWriter, request *http.Request) {
	if request.FormValue("load_all") == "1" {
		err := app.loadAllModels()
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
	}
	modelList := []*Model{}
	for _, model := range app.models {
		modelList = append(modelList, model)
	}
	jsonModelList, _ := json.Marshal(modelList)
	fmt.Fprintf(response, string(jsonModelList))
}

/*
   trainModel displays a form for training a model with an new observation and handles
   form submission, training the given model
   * GET /model/<name>/train - Displays model train form
   * POST /model/<name/train - Trains the given model with the input observation
*/
func (app *NaiveBayesApp) trainModel(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	modelName := vars["modelName"]
	model, err := app.loadModel(modelName)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
	if request.Method == "GET" {
		app.renderTemplate(response, "model_train.html", model)
	} else if request.Method == "POST" {
		request.ParseForm()
		observation_classes_str := request.FormValue("observation_classes")
		observation_text := request.FormValue("observation_text")
		observation_classes := strings.Split(observation_classes_str, ",")
		observation := NewObservationFromText(observation_classes, observation_text)
		model.Train(observation)
		log.Printf("Trained model: '%s' with new observation for classes: '%s'", model.Name, observation_classes)
		SaveToFile(app.modelPath(model.Name), model, json.Marshal)
		http.Redirect(response, request, "/model/"+modelName, 301)
	}
}

/*
   predictModel displays a form for predicting classes
   for a new observation based on the given model
   and handles form submission returning the results of the prediction in JSON
   * GET /model/<name>/predict - Displays model prediction form
   * POST /model/<name/predict - Predicts the class for the input using the given model
*/
func (app *NaiveBayesApp) predictModel(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	modelName := vars["modelName"]
	model, err := app.loadModel(modelName)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
	if request.Method == "GET" {
		app.renderTemplate(response, "model_predict.html", model)
	} else if request.Method == "POST" {
		request.ParseForm()
		predict_text := request.FormValue("predict_text")
		observation := NewObservationFromText([]string{}, predict_text)
		prediction := model.Predict(observation)
		jsonPrediction, _ := json.Marshal(prediction)
		fmt.Fprintf(response, string(jsonPrediction))
	}
}
