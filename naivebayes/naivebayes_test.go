package naivebayes

import (
	"fmt"
	"testing"
)

var model *Model

// roundProbability rounds probabilities to avoid intermittent failures
func roundProbability(p float64) (s string) {
	return fmt.Sprintf("%.15f", s)
}

// TestModel is an overall test
// Creates a new model, trains it, and checks predicted probablities
func TestModel(t *testing.T) {
	model = NewModel("hello")

	obs1 := NewObservationFromText([]string{"China"}, "Chinese Beijing Chinese")
	model.Train(obs1)

	obs2 := NewObservationFromText([]string{"China"}, "Chinese Chinese Shanghai")
	model.Train(obs2)

	obs3 := NewObservationFromText([]string{"China"}, "Chinese Macao")
	model.Train(obs3)

	obs4 := NewObservationFromText([]string{"NotChina"}, "Tokyo Japan Chinese")
	model.Train(obs4)

	testObs := NewObservationFromText([]string{}, "Chinese Chinese Chinese Tokyo Japan")

	expectedChina := roundProbability(0.00030121377997263)
	expectedNotChina := roundProbability(0.00013548070246744215)

	predictions := model.Predict(testObs)

	if roundProbability(predictions["China"]) != expectedChina {
		t.Errorf("Did not get expected probability for China. Expected: %d, Got: %d", expectedChina, predictions["China"])
	}

	if roundProbability(predictions["NotChina"]) != expectedNotChina {
		t.Errorf("Did not get expected probability for NotChina. Expected: %d, Got: %d", expectedNotChina, predictions["NotChina"])
	}

	name, value := predictions.BestFit()

	if name != "China" {
		t.Error("Did not predict best fit name")
	}

	if roundProbability(value) != expectedChina {
		t.Error("Did not predict best fit value")
	}
}
