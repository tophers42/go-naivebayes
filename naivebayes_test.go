package naivebayes

import "testing"

func TestModelFulle(t *testing.T) {
	model := NewModel("hello")

	obs1 := NewObservationFromText([]string{"China"}, "Chinese Beijing Chinese")
	model.Train(obs1)

	obs2 := NewObservationFromText([]string{"China"}, "Chinese Chinese Shanghai")
	model.Train(obs2)

	obs3 := NewObservationFromText([]string{"China"}, "Chinese Macao")
	model.Train(obs3)

	obs4 := NewObservationFromText([]string{"NotChina"}, "Tokyo Japan Chinese")
	model.Train(obs4)

	testObs := NewObservationFromText([]string{}, "Chinese Chinese Chinese Tokyo Japan")

	expectedChina := 0.00030121377997263
	expectedNotChina := 0.00013548070246744215

	predictions := model.predict(testObs)

	if predictions["China"] != expectedChina {
		t.Errorf("Did not get expected probability for China. Expected: %d, Got: %d", expectedChina, predictions["China"])
	}

	if predictions["NotChina"] != expectedNotChina {
		t.Errorf("Did not get expected probability for NotChina. Expected: %d, Got: %d", expectedNotChina, predictions["NotChina"])
	}

}
