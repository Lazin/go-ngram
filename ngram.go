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
	count int64
}

type NGramIndex struct {
	pad   string
	n     int
	spool stringPool
	index map[uint32]*nGramValue
}

type SearchResult struct {
	TokenId    TokenId
	Similarity float32
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
		val.count++
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
	counters := make(map[TokenId]int)
	for _, ngram_hash := range input_ngrams {
		if tokmap, ok := ngram.index[ngram_hash]; ok {
			for tok, _ := range tokmap.items {
				counters[tok] += 1
			}
		}
	}
	return counters
}

func (ngram *NGramIndex) Search(input string, threshold float32) ([]SearchResult, error) {
	if ngram.index == nil {
		ngram.init()
	}
	if threshold < 0.0 || threshold > 1.0 {
		return nil, errors.New("Threshold must be in range (0, 1)")
	}
	input_ngrams, error := ngram.split_input(input)
	if error != nil {
		return nil, error
	}
	output := make([]SearchResult, 0)
	token_count := ngram.count_ngrams(input_ngrams)
	for token, count := range token_count {
		sim := float32(count) / float32(len(input_ngrams))
		if sim >= threshold {
			res := SearchResult{Similarity: sim, TokenId: token}
			output = append(output, res)
		}
	}
	return output, nil
}
