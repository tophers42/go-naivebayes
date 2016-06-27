package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/tophers42/go-naivebayes/naivebayes"
)

type App struct {
	templates *template.Template
	modelDir  string
	models    map[string]*naivebayes.Model
}

func newApp() (app *App) {
	// init a model storage dir
	modelDir := "saved_models"
	os.Mkdir(modelDir, 0644)

	// app.validModelName = regexp.MustCompile("([a-zA-Z0-9]+)$")

	// init the templates
	var templates = template.Must(template.ParseFiles("templates/model_create.html", "templates/model_predict.html"))
	fmt.Println(templates.DefinedTemplates())

	app = &App{templates: templates, models: make(map[string]*naivebayes.Model), modelDir: modelDir}
	return app
}

func (app *App) modelPath(modelName string) (path string) {
	return app.modelDir + "/" + modelName + ".json"
}

func (app *App) loadModel(modelName string) (model *naivebayes.Model, err error) {
	model, ok := app.models[modelName]
	if ok {
		return model, nil
	}
	log.Printf("Model: %s not found in memory. Falling back to file", modelName)
	model, err = naivebayes.NewModelFromFile(app.modelPath(modelName))
	if err == nil {
		return model, nil
	}
	log.Printf("Failed to load model from file: %s, err: %s", modelName, err)
	return nil, fmt.Errorf("Unable to load model: %s", modelName)
}

func (app *App) makeHandler(handlerFunc func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		log.Printf("Handling request to: %s", request.URL.Path)
		handlerFunc(response, request)
	}
}

func (app *App) renderTemplate(response http.ResponseWriter, tmpl string, data interface{}) {
	err := app.templates.ExecuteTemplate(response, tmpl, data)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func (app *App) modelView(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	modelName := vars["modelName"]
	model, err := app.loadModel(modelName)

	if err == nil {
		jsonModel, _ := json.Marshal(model)
		fmt.Fprintf(response, string(jsonModel))
	} else {
		http.Error(response, fmt.Sprintf("Unable to load model by the name of: %s", modelName), http.StatusNotFound)
	}
}

func (app *App) modelList(response http.ResponseWriter, request *http.Request) {
	var modelList []*naivebayes.Model
	for _, model := range app.models {
		modelList = append(modelList, model)
	}
	jsonModelList, _ := json.Marshal(modelList)
	fmt.Fprintf(response, string(jsonModelList))
}

func (app *App) modelCreate(response http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		app.renderTemplate(response, "model_create.html", nil)
	} else if request.Method == "POST" {
		request.ParseForm()
		modelName := request.FormValue("model_name")
		modelExists, _ := app.loadModel(modelName)
		if modelExists != nil {
			fmt.Fprintf(response, "Model by that name already exists: %s", modelName)
		} else {
			model := naivebayes.NewModel(modelName)
			err := model.SaveToFile("saved_models/" + modelName + ".json")
			if err != nil {
				http.Error(response, err.Error(), http.StatusInternalServerError)
			} else {
				log.Printf("Created new model: %s", modelName)

				app.models[modelName] = model
				http.Redirect(response, request, "model/"+modelName, 200)
			}
		}
	}
}

// endpoints:
// GET model/<name> - view model
// GET model - new model form
// POST model - new model
// GET model/<name>/train - train model form
// POST model/<name>/train - train model with new observation
// GET model/<name>/predict - predict form
// POST model/<name>/predict - predict observation using model
func (app *App) registerEndpoints() {
	router := mux.NewRouter()
	router.HandleFunc("/model", app.makeHandler(app.modelCreate)).Methods("GET", "POST")
	router.HandleFunc("/models", app.makeHandler(app.modelList)).Methods("GET")
	router.HandleFunc("/model/{modelName}", app.makeHandler(app.modelView)).Methods("GET")
	// router.HandleFunc("/model/{modelName}/train", app.makeHandler(app.modelTrain)).Methods("GET", "POST")
	// router.HandleFunc("/model/{modelName}/predict", app.makeHandler(app.modelPredict)).Methods("GET", "POST")
	http.Handle("/", router)
}

func main() {
	app := newApp()

	app.registerEndpoints()
	log.Print("listening")
	http.ListenAndServe(":8080", nil)

}
