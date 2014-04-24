package ngram

import (
	"errors"
	"github.com/reusee/mmh3"
	"math"
)

const maxN = 8

const defaultPad = "$"

const defaultN = 3

type TokenId int

type nGramValue struct {
	items map[TokenId]int
	count int64
}

// N-gram index, can be initialized by default (zeroed) or created with "NewNgramIndex"
type NGramIndex struct {
	pad   string
	n     int
	spool stringPool
	index map[uint32]*nGramValue
	warp  float64
}

// Search result, contains token id and similarity - value in range from 0.0 to 1.0
type SearchResult struct {
	TokenId    TokenId
	Similarity float64
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
	if ngram.warp == 0.0 {
		ngram.warp = 1.0
	}
}

type padArgTrait struct {
	pad rune
}

type nArgTrait struct {
	n int
}

type warpArgTrait struct {
	warp float64
}

// This function must be used to pass padding character to NGramIndex c-tor
func SetPad(c rune) padArgTrait {
	return padArgTrait{pad: c}
}

// This function must be used to pass N (gram size) to NGramIndex c-tor
func SetN(n int) nArgTrait {
	return nArgTrait{n: n}
}

// This function must be used to pass warp to NGramIndex c-tor
func SetWarp(warp float64) warpArgTrait {
	return warpArgTrait{warp: warp}
}

// N-gram index c-tor. In most cases must be used withot parameters.
// You can pass parameters to c-tor using functions SetPad and SetN.
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
		case warpArgTrait:
			if i.warp < 0.0 || i.warp > 1.0 {
				return nil, errors.New("Bad 'warp' value for n-gram index")
			}
		default:
			return nil, errors.New("Invalid argument")
		}
	}
	ngram.init()
	return ngram, nil
}

// Add token to index. Function returns token id, this id can be converted
// to string with function "GetString".
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

// Converts token-id to string.
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

// Search for matches between query string (input) and indexed strings.
// First parameter - threshold is optional and can be used to set minimal similarity
// between input string and matching string. You can pass only one threshold value.
// Results is an unordered array of 'SearchResult' structs. This struct contains similarity
// value (float32 value from threshold to 1.0) and token-id.
func (ngram *NGramIndex) Search(input string, threshold ...float64) ([]SearchResult, error) {
	if ngram.index == nil {
		ngram.init()
	}
	var threshold_val float64
	if len(threshold) == 1 {
		threshold_val = threshold[0]
		if threshold_val < 0.0 || threshold_val > 1.0 {
			return nil, errors.New("Threshold must be in range (0, 1)")
		}
	} else if len(threshold) > 1 {
		return nil, errors.New("Too many arguments")
	}
	input_ngrams, error := ngram.split_input(input)
	if error != nil {
		return nil, error
	}
	output := make([]SearchResult, 0)
	token_count := ngram.count_ngrams(input_ngrams)
	for token, count := range token_count {
		var sim float64
		allngrams := float64(len(input_ngrams))
		matchngrams := float64(count)
		if ngram.warp == 1.0 {
			sim = matchngrams / allngrams
		} else {
			diffngrams := allngrams - matchngrams
			sim = math.Pow(allngrams, ngram.warp) - math.Pow(diffngrams, ngram.warp)
			sim /= math.Pow(allngrams, ngram.warp)
		}
		if sim >= threshold_val {
			res := SearchResult{Similarity: sim, TokenId: token}
			output = append(output, res)
		}
	}
	return output, nil
}
