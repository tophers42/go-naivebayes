package naivebayes

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestModelTrain(t *testing.T) {
	model := NewModel("hello")
	obs := Observation{Categories: []Category{"a", "b"}, Content: []string{"hello", "bye", "bye", "hellohello"}}

	model.Train(obs)
	if model.Counts["hello"].Counts["a"] != 1 {
		t.Error()
	}

	if model.Counts["hello"].Counts["b"] != 1 {
		t.Error()
	}

	if model.Counts["bye"].Counts["a"] != 2 {
		t.Error()
	}

	model.Train(obs)

	if model.Counts["hello"].Counts["a"] != 2 {
		t.Error()
	}

	j, _ := json.Marshal(model)
	fmt.Println(string(j))

}
