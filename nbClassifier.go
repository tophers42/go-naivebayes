// nbClassifier
package naivebayes

/*
TODO:
	* rearrange/rename
	* cleanup comments
	* expand tests
	* saveable/loadable
	* wrapper web server
	* dockerize
	* smarter text parsing ( stemming, synonyms )
*/

import (
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
	observationCount int
	wordCounts       map[string]int
	totalCount       int
}

func NewClass(name string) *Class {
	return &Class{Name: name, wordCounts: make(map[string]int), totalCount: 0}
}

func (c *Class) addWord(word string, count int) {
	c.wordCounts[word] += count
	c.totalCount += count
}

/*
Model struct
*/
type Model struct {
	Name             string
	classes          map[string]*Class
	observationCount int
	vocabulary       map[string]int
}

func NewModel(name string) *Model {
	return &Model{Name: name, classes: make(map[string]*Class), observationCount: 0, vocabulary: make(map[string]int)}
}

/*
 P( class ) - prior probability of this class within the model
 Number of occurences of the class / Number of trained observations

*/

func (m *Model) classPriorProbability(class *Class) (p float64) {
	p = math.Log(float64(class.observationCount) / float64(m.observationCount))
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
		classWordCount, _ := class.wordCounts[word]
		raw := float64(classWordCount+1) / float64(class.totalCount+len(m.vocabulary))
		p = p + (math.Log(raw) * float64(count))
	}
	return p
}

// train the model with the given observation
func (m *Model) Train(o *Observation) {
	for _, className := range o.Classes {
		class, ok := m.classes[className]
		if !ok {
			class = NewClass(className)
			m.classes[className] = class
		}
		class.observationCount++
		for word, count := range o.WordCounts {
			class.addWord(word, count)
			m.vocabulary[word] = 1
		}
	}

	m.observationCount++
}

func (m *Model) predict(o *Observation) (p Prediction) {
	p = make(map[string]float64)

	for _, class := range m.classes {
		classConditionalProbability := m.classPriorProbability(class) + m.classConditionalProbability(class, o)
		p[class.Name] = math.Exp(classConditionalProbability)
	}
	return p
}
