// nbClassifier
package naivebayes

type Category string

type Prediction map[Category]float64

type Observation struct {
	Categories []Category
	Content    []string
}

/*
Model struct
*/
type WordCount struct {
	Word   string
	Counts map[Category]int
}

func NewWordCount(word string) *WordCount {
	return &WordCount{Word: word, Counts: make(map[Category]int)}
}

// func (wc *WordCount) AddObservation(o *Observation) {

// }

/*
Model struct
*/
type Model struct {
	Name   string                `json:"name"`
	Counts map[string]*WordCount `json:"counts"`
}

func NewModel(name string) *Model {
	return &Model{Name: name, Counts: make(map[string]*WordCount)}
}

// train the model with the given observation
func (m *Model) Train(o Observation) {
	for _, word := range o.Content {
		count, ok := m.Counts[word]
		if !ok {
			count = NewWordCount(word)
			m.Counts[word] = count
		}

		for _, category := range o.Categories {
			count.Counts[category]++
		}
	}
}

//
// func (m *Model) predict( o Observation ) p *Prediction {

// }
