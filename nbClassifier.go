// nbClassifier
package naivebayes

import (
	"fmt"
	"math"
	"strings"
)

type Observation struct {
	Classes    []string
	WordCounts map[string]int
}

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

type Prediction map[string]float64

//TODO: FIX VOCAB COUNT to be model level, not per class.
// http://nlp.stanford.edu/IR-book/html/htmledition/naive-bayes-text-classification-1.html

/*
Class struct
*/
type Class struct {
	Name             string
	WordCounts       map[string]int
	VocabCount       int
	TotalCount       int
	ObservationCount int
}

func NewClass(name string) *Class {
	return &Class{Name: name, WordCounts: make(map[string]int), ObservationCount: 0, TotalCount: 0}
}

func (c *Class) addObservation(o *Observation) {
	c.ObservationCount++
	for word, count := range o.WordCounts {
		_, ok := c.WordCounts[word]
		if !ok {
			// new word
			c.VocabCount++
		}
		c.WordCounts[word] += count
		c.TotalCount += count
	}
}

/*
 P( observation | class ) - conditional probability of this observation
 given the class.
 Multiply the probabilities of each word being in this class

*/

func (c *Class) observationConditionalProbability(o *Observation) (p float64) {
	p = 0
	for word, count := range o.WordCounts {
		p = p + c.wordConditionalProbability(word, count)
	}
	return p

}

/*
 P( word | class ) - probability of a word, given a class
 Occurences of the word within the class, divided by total number of words
 Laplace smoothing (always add one so new words don't break everything)
 and also add unique word count to denominator
*/

func (c *Class) wordConditionalProbability(word string, count int) (p float64) {
	classWordCount, _ := c.WordCounts[word]
	raw := float64(classWordCount+1) / float64(c.TotalCount+c.VocabCount)
	p = math.Log(raw) * float64(count)
	fmt.Printf("Word: %s, Count: %d, Class: %s, Raw: %f, Prediction: %f\n", word, count, c.Name, raw, p)
	return p
}

/*
Model struct
*/
type Model struct {
	Name             string `json:"name"`
	Classes          map[string]*Class
	ObservationCount int
}

func NewModel(name string) *Model {
	return &Model{Name: name, Classes: make(map[string]*Class), ObservationCount: 0}
}

/*
 P( class ) - prior probability of this class within the model
 Number of occurences of the class / Number of trained observations

*/

func (m *Model) classPriorProbability(className string) (p float64) {
	// assume 0 if this class is unknown
	p = 0
	class, ok := m.Classes[className]
	if ok {
		p = math.Log(float64(class.ObservationCount) / float64(m.ObservationCount))
	}
	return p

}

// train the model with the given observation
func (m *Model) Train(o *Observation) {
	for _, class_name := range o.Classes {
		class, ok := m.Classes[class_name]
		if !ok {
			class = NewClass(class_name)
			m.Classes[class_name] = class
		}
		class.addObservation(o)
	}
	m.ObservationCount++
}

func (m *Model) predict(o *Observation) (p Prediction) {
	p = make(map[string]float64)

	for className, class := range m.Classes {
		classConditionalProbability := m.classPriorProbability(className) + class.observationConditionalProbability(o)
		p[className] = math.Exp(classConditionalProbability)
	}
	return p
}
