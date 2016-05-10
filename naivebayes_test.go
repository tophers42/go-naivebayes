package naivebayes

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestModelTrain(t *testing.T) {
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

	spew.Dump(model.Classes)

	fmt.Println(model.predict(testObs))

}
