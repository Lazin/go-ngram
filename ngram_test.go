package ngram

import "testing"

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
