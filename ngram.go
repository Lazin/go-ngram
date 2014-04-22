package ngram

import (
	"errors"
	"github.com/reusee/mmh3"
)

const maxN = 8

const defaultPad = "$"

const defaultN = 3

type TokenId int

type nGramValue struct {
	items map[TokenId]int
}

type NGramIndex struct {
	pad   string
	n     int
	spool stringPool
	index map[uint32]*nGramValue
}

type SearchResult struct {
	text       string
	tokenId    TokenId
	similarity float32
	error      error
}

func (ngram *NGramIndex) split_input(str string) ([]uint32, error) {
	if len(str) == 0 {
		return nil, errors.New("Empty string")
	}
	pad := ngram.pad
	n := ngram.n
	input := pad + str + pad
	prev_indexes := make([]int, maxN)
	counter := 0
	results := make([]uint32, 0)

	for index, _ := range input {
		counter++
		if counter > n {
			top := prev_indexes[(counter-n)%len(prev_indexes)]
			substr := input[top:index]
			hash := mmh3.Hash32([]byte(substr))
			results = append(results, hash)
		}
		prev_indexes[counter%len(prev_indexes)] = index
	}

	for i := n - 1; i > 1; i-- {
		if len(input) >= i {
			top := prev_indexes[(len(input)-i)%len(prev_indexes)]
			substr := input[top:]
			hash := mmh3.Hash32([]byte(substr))
			results = append(results, hash)
		}
	}

	return results, nil
}

func (ngram *NGramIndex) init() {
	ngram.index = make(map[uint32]*nGramValue)
	if ngram.pad == "" {
		ngram.pad = defaultPad
	}
	if ngram.n == 0 {
		ngram.n = defaultN
	}
}

type padArgTrait struct {
	pad rune
}

type nArgTrait struct {
	n int
}

func SetPad(c rune) padArgTrait {
	return padArgTrait{pad: c}
}

func SetN(n int) nArgTrait {
	return nArgTrait{n: n}
}

func NewNGramIndex(args ...interface{}) (*NGramIndex, error) {
	ngram := new(NGramIndex)
	for _, arg := range args {
		switch i := arg.(type) {
		case padArgTrait:
			ngram.pad = string(i.pad)
		case nArgTrait:
			if i.n < 2 || i.n > maxN {
				return nil, errors.New("Bad 'n' value for n-gram index")
			}
			ngram.n = i.n
		default:
			return nil, errors.New("Invalid argument")
		}
	}
	ngram.init()
	return ngram, nil
}

func (ngram *NGramIndex) Add(input string) (TokenId, error) {
	if ngram.index == nil {
		ngram.init()
	}
	results, error := ngram.split_input(input)
	if error != nil {
		return -1, error
	}
	ixstr, error := ngram.spool.Append(input)
	if error != nil {
		return -1, error
	}
	for _, hash := range results {
		var val *nGramValue = nil
		var ok bool
		if val, ok = ngram.index[hash]; !ok {
			val = new(nGramValue)
			val.items = make(map[TokenId]int)
			ngram.index[hash] = val
		}
		// insert string and counter
		if count, ok := val.items[ixstr]; ok {
			val.items[ixstr] = count + 1
		} else {
			val.items[ixstr] = 1
		}
	}
	return ixstr, nil
}

func (ngram *NGramIndex) GetString(id TokenId) (string, error) {
	return ngram.spool.ReadAt(id)
}

// Map matched tokens to the number of ngrams, shared with input string
func (ngram *NGramIndex) count_ngrams(input_ngrams []uint32) map[TokenId]int {
	panic("Not implemented")
}

// Return best match
func (ngram *NGramIndex) Find(input string) (TokenId, error) {
	if ngram.index == nil {
		ngram.init()
	}
	panic("Not implemented")
}

func (ngram *NGramIndex) Search(input string) ([]SearchResult, error) {
	if ngram.index == nil {
		ngram.init()
	}
	results, error := ngram.split_input(input)
	if error != nil {
		return nil, error
	}
	output := make([]SearchResult, 0)
	for _, hash := range results {
		if val := ngram.index[hash]; val != nil {
			for id, _ := range val.items {
				text, err := ngram.spool.ReadAt(id)
				res := SearchResult{text: text, error: err, similarity: 0.0, tokenId: id}
				output = append(output, res)
			}
		}
	}
	return output, nil
}
