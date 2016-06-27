package naivebayes

/*
NaiveBayes text classifier.
Based on: http://nlp.stanford.edu/IR-book/html/htmledition/naive-bayes-text-classification-1.html

TODO:
	* rearrange/rename
	* cleanup comments
	* expand tests
	* wrapper web server
	* dockerize
	* smarter text parsing ( stemming, synonyms )
*/

import (
	"encoding/json"
	"io/ioutil"
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

func (p *Prediction) BestFit() (bestClassName string, bestProbability float64) {
	for className, probability := range *p {
		if bestClassName == "" || probability > bestProbability {
			bestClassName = className
			bestProbability = probability
		}
	}
	return bestClassName, bestProbability
}

/*
Class struct
*/
type Class struct {
	Name             string
	ObservationCount int
	WordCounts       map[string]int
	TotalCount       int
}

// Build an empty class struct
func NewClass(name string) *Class {
	return &Class{Name: name, WordCounts: make(map[string]int), TotalCount: 0}
}

// Increment the count for the given word on the class
func (c *Class) addWord(word string, count int) {
	c.WordCounts[word] += count
	c.TotalCount += count
}

/*
Model struct
*/
type Model struct {
	Name             string
	Classes          map[string]*Class
	ObservationCount int
	Vocabulary       map[string]int
}

func NewModel(name string) *Model {
	return &Model{Name: name, Classes: make(map[string]*Class), ObservationCount: 0, Vocabulary: make(map[string]int)}
}

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

func (m *Model) SaveToFile(path string) (err error) {
	jsonModel, _ := json.Marshal(m)
	// write the whole thing at once
	err = ioutil.WriteFile(path, jsonModel, 0644)
	return err
}

/*
 P( class ) - prior probability of this class within the model
 Number of occurences of the class / Number of trained observations
*/
func (m *Model) classPriorProbability(class *Class) (p float64) {
	p = math.Log(float64(class.ObservationCount) / float64(m.ObservationCount))
	return p

}

/*
 P( observation | class ) - conditional probability of this observation
 given the class.
 Multiply the probabilities of each word being in this class (sum logs)
 P( word | class ) - probability of a word, given a class
 Occurences of the word within the class, divided by total number of words
 Laplace smoothing (always add one so new words don't break everything)
 and also add unique word count to denominator
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

// train the model with the given observation
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

// calculate observation membership probability for each class
func (m *Model) Predict(o *Observation) (p Prediction) {
	p = make(map[string]float64)

	for _, class := range m.Classes {
		classConditionalProbability := m.classPriorProbability(class) + m.classConditionalProbability(class, o)
		p[class.Name] = math.Exp(classConditionalProbability)
	}
	return p
}
