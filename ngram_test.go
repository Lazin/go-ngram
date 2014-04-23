package ngram

import (
	"testing"
)

func Test_index_basics(t *testing.T) {
	var ng NGramIndex
	index := &ng
	id, error := index.Add("hello")
	if error != nil {
		t.Error(error)
	}
	strval, error := index.GetString(id)
	if error != nil {
		t.Error(error)
	}
	if strval != "hello" {
		t.Error("Can't read string from index")
	}
}

func Test_searching(t *testing.T) {
	var ng NGramIndex
	index := &ng
	_, error := index.Add("hello")
	if error != nil {
		t.Error(error)
	}
	_, error = index.Add("world")
	if error != nil {
		t.Error(error)
	}
	results, error := index.Search("hello", 0.0)
	if error != nil {
		t.Error(error)
	}
	if len(results) != 1 {
		t.Error("len(results) != 1")
	}
	if results[0].Similarity != 1.0 && results[0].TokenId != 0 {
		t.Error("Bad result")
	}
	results, error = index.Search("12345")
	if len(results) != 0 {
		t.Error("Invalid value found")
	}
}

func Test_index_initialization(t *testing.T) {
	index, error := NewNGramIndex()
	if error != nil {
		t.Error(error)
	}
	if index.n != defaultN {
		t.Error("n is not set to default value")
	}
	if index.pad != defaultPad {
		t.Error("pad is not set to default value")
	}
	index, error = NewNGramIndex(SetN(4))
	if error != nil {
		t.Error(error)
	}
	if index.n != 4 {
		t.Error("n is not set to 4")
	}
	index, error = NewNGramIndex(SetPad('@'))
	if error != nil {
		t.Error(error)
	}
	if index.pad != "@" {
		t.Error("pad is not set to @")
	}
	// check off limits
	index, error = NewNGramIndex(SetN(1))
	if error == nil {
		t.Error("Error not set (1)")
	}
	index, error = NewNGramIndex(SetN(maxN + 1))
	if error == nil {
		t.Error("Error not set (2)")
	}
}
