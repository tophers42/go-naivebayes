// Package naivebayes provides functionality and structs for building and training a naivebayes
// text classifier  model and making predictions based on that model.
package naivebayes

/*
NaiveBayes text classifier.
Based on: http://nlp.stanford.edu/IR-book/html/htmledition/naive-bayes-text-classification-1.html

TODO:
	* smarter text parsing ( stemming, synonyms )
*/

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"strings"
)

// Observation struct.
// Represents an instance of text to be classified (or for training).
type Observation struct {
	Classes    []string
	WordCounts map[string]int
}

// NewObservationFromText creates an observation object.
// Breaks up a block of text on " " and calculates word counts.
func NewObservationFromText(classes []string, text string) *Observation {
	counts := make(map[string]int)
	for _, word := range strings.Split(text, " ") {
		_, ok := counts[word]
		if !ok {
			counts[word] = 0
		}
		counts[word]++
	}
	return &Observation{Classes: classes, WordCounts: counts}
}

// Prediction type. A mapping of class names to calculated probabilities.
// Returned when calling Predict on a trained model with a new observation.
type Prediction map[string]float64

// BestFit filters a prediction down to the best fit,
// returning just the class name and calulated probability.
func (p *Prediction) BestFit() (bestClassName string, bestProbability float64) {
	for className, probability := range *p {
		if bestClassName == "" || probability > bestProbability {
			bestClassName = className
			bestProbability = probability
		}
	}
	return bestClassName, bestProbability
}

// Class struct (a.k.a category).
// Represents a grouping of observations that belong together.
type Class struct {
	Name             string
	ObservationCount int
	WordCounts       map[string]int
	TotalCount       int
}

// NewClasse creates an empty class struct.
func NewClass(name string) *Class {
	return &Class{Name: name, WordCounts: make(map[string]int), TotalCount: 0}
}

// addWord increments the count for the given word on the Class.
func (c *Class) addWord(word string, count int) {
	c.WordCounts[word] += count
	c.TotalCount += count
}

// Model struct.
// Represents a training set and can be used to make class membership
// predictions on new observations.
type Model struct {
	Name             string
	Classes          map[string]*Class
	ObservationCount int
	Vocabulary       map[string]int
}

// NewModel creates and empty Model with the given name.
func NewModel(name string) *Model {
	return &Model{Name: name, Classes: make(map[string]*Class), ObservationCount: 0, Vocabulary: make(map[string]int)}
}

// NewModelFromFile creates a new Model and loads its state from the given file.
// (Model state is stored as a simple json dump of the Model struct).
func NewModelFromFile(path string) (m *Model, err error) {
	// read the whole thing at once
	m = &Model{}
	jsonModel, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonModel, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// SaveToFile dumps the current state of the Model to the given file.
// (Model state is stored as a simple json dump of the Model struct).
func (m *Model) SaveToFile(path string) (err error) {
	jsonModel, _ := json.Marshal(m)
	// write the whole thing at once
	err = ioutil.WriteFile(path, jsonModel, 0644)
	return err
}

// Train updates (trains) the Model with the given Observation.
func (m *Model) Train(o *Observation) {
	for _, className := range o.Classes {
		class, ok := m.Classes[className]
		if !ok {
			class = NewClass(className)
			m.Classes[className] = class
		}
		class.ObservationCount++
		for word, count := range o.WordCounts {
			class.addWord(word, count)
			m.Vocabulary[word] = 1
		}
	}

	m.ObservationCount++
}

// Predict caalculates the posterior probabality for the given observation
// for each class within the Model.
func (m *Model) Predict(o *Observation) (p Prediction) {
	p = make(map[string]float64)

	for _, class := range m.Classes {
		classConditionalProbability := m.classPriorProbability(class) + m.classConditionalProbability(class, o)
		p[class.Name] = math.Exp(classConditionalProbability)
	}
	return p
}

// classPriorProbability calculates the prior probability of the given class
// for the Model.
// P( class ) - Number of occurences of the class / Number of trained observations
func (m *Model) classPriorProbability(class *Class) (p float64) {
	p = math.Log(float64(class.ObservationCount) / float64(m.ObservationCount))
	return p

}

/*
	classConditionalProbability calculates the conditional probability of the
	given observation for the given class within the Model.
	P( observation | class ) - conditional probability of this observation
	given the class.

	Multiplies the probabilities of each word being in this class (sum logs).
	P( word | class ) - probability of a word, given a class
	Occurences of the word within the class, divided by total number of words
	(words in class plus total number of unique words seen by the model)

	Laplace smoothing (always add one so new words don't break everything)
	and also add unique word count to denominator.
*/
func (m *Model) classConditionalProbability(class *Class, o *Observation) (p float64) {
	p = 0
	for word, count := range o.WordCounts {
		classWordCount, _ := class.WordCounts[word]
		raw := float64(classWordCount+1) / float64(class.TotalCount+len(m.Vocabulary))
		p = p + (math.Log(raw) * float64(count))
	}
	return p
}
