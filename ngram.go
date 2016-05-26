package ngram

import (
	"errors"
	"math"
	"sync"

	"github.com/spaolacci/murmur3"
)

const (
	maxN       = 8
	defaultPad = "$"
	defaultN   = 3
)

// TokenID is just id of the token
type TokenID int

type nGramValue map[TokenID]int

// NGramIndex can be initialized by default (zeroed) or created with "NewNgramIndex"
type NGramIndex struct {
	pad   string
	n     int
	spool stringPool
	index map[uint32]nGramValue
	warp  float64

	sync.RWMutex
}

// SearchResult contains token id and similarity - value in range from 0.0 to 1.0
type SearchResult struct {
	TokenID    TokenID
	Similarity float64
}

func (ngram *NGramIndex) splitInput(str string) ([]uint32, error) {
	if len(str) == 0 {
		return nil, errors.New("empty string")
	}
	pad := ngram.pad
	n := ngram.n
	input := pad + str + pad
	prevIndexes := make([]int, maxN)
	counter := 0
	results := make([]uint32, 0)

	for index := range input {
		counter++
		if counter > n {
			top := prevIndexes[(counter-n)%len(prevIndexes)]
			substr := input[top:index]
			hash := murmur3.Sum32([]byte(substr))
			results = append(results, hash)
		}
		prevIndexes[counter%len(prevIndexes)] = index
	}

	for i := n - 1; i > 1; i-- {
		if len(input) >= i {
			top := prevIndexes[(len(input)-i)%len(prevIndexes)]
			substr := input[top:]
			hash := murmur3.Sum32([]byte(substr))
			results = append(results, hash)
		}
	}

	return results, nil
}

func (ngram *NGramIndex) init() {
	ngram.Lock()
	defer ngram.Unlock()

	ngram.index = make(map[uint32]nGramValue)
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

type Option func(*NGramIndex) error

// SetPad must be used to pass padding character to NGramIndex c-tor
func SetPad(c rune) Option {
	return func(ngram *NGramIndex) error {
		ngram.pad = string(c)
		return nil
	}
}

// SetN must be used to pass N (gram size) to NGramIndex c-tor
func SetN(n int) Option {
	return func(ngram *NGramIndex) error {
		if n < 2 || n > maxN {
			return errors.New("bad 'n' value for n-gram index")
		}
		ngram.n = n
		return nil
	}
}

// SetWarp must be used to pass warp to NGramIndex c-tor
func SetWarp(warp float64) Option {
	return func(ngram *NGramIndex) error {
		if warp < 0.0 || warp > 1.0 {
			return errors.New("bad 'warp' value for n-gram index")
		}
		ngram.warp = warp
		return nil
	}
}

// NewNGramIndex is N-gram index c-tor. In most cases must be used withot parameters.
// You can pass parameters to c-tor using functions SetPad, SetWarp and SetN.
func NewNGramIndex(opts ...Option) (*NGramIndex, error) {
	ngram := new(NGramIndex)
	for _, opt := range opts {
		if err := opt(ngram); err != nil {
			return nil, err
		}
	}
	ngram.init()
	return ngram, nil
}

// Add token to index. Function returns token id, this id can be converted
// to string with function "GetString".
func (ngram *NGramIndex) Add(input string) (TokenID, error) {
	if ngram.index == nil {
		ngram.init()
	}
	results, error := ngram.splitInput(input)
	if error != nil {
		return -1, error
	}
	ixstr, error := ngram.spool.Append(input)
	if error != nil {
		return -1, error
	}
	for _, hash := range results {
		ngram.Lock()
		if ngram.index[hash] == nil {
			ngram.index[hash] = make(map[TokenID]int)
		}
		// insert string and counter
		ngram.index[hash][ixstr]++
		ngram.Unlock()
	}
	return ixstr, nil
}

// GetString converts token-id to string.
func (ngram *NGramIndex) GetString(id TokenID) (string, error) {
	return ngram.spool.ReadAt(id)
}

// countNgrams maps matched tokens to the number of ngrams, shared with input string
func (ngram *NGramIndex) countNgrams(inputNgrams []uint32) map[TokenID]int {
	counters := make(map[TokenID]int)
	for _, ngramHash := range inputNgrams {
		ngram.RLock()
		for tok := range ngram.index[ngramHash] {
			counters[tok]++
		}
		ngram.RUnlock()
	}
	return counters
}

func validateThresholdValues(thresholds []float64) (float64, error) {
	var tval float64
	if len(thresholds) == 1 {
		tval = thresholds[0]
		if tval < 0.0 || tval > 1.0 {
			return 0.0, errors.New("threshold must be in range (0, 1)")
		}
	} else if len(thresholds) > 1 {
		return 0.0, errors.New("too many arguments")
	}
	return tval, nil
}

func (ngram *NGramIndex) match(input string, tval float64) ([]SearchResult, error) {
	inputNgrams, error := ngram.splitInput(input)
	if error != nil {
		return nil, error
	}
	output := make([]SearchResult, 0)
	tokenCount := ngram.countNgrams(inputNgrams)
	for token, count := range tokenCount {
		var sim float64
		allngrams := float64(len(inputNgrams))
		matchngrams := float64(count)
		if ngram.warp == 1.0 {
			sim = matchngrams / allngrams
		} else {
			diffngrams := allngrams - matchngrams
			sim = math.Pow(allngrams, ngram.warp) - math.Pow(diffngrams, ngram.warp)
			sim /= math.Pow(allngrams, ngram.warp)
		}
		if sim >= tval {
			res := SearchResult{Similarity: sim, TokenID: token}
			output = append(output, res)
		}
	}
	return output, nil
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
	tval, error := validateThresholdValues(threshold)
	if error != nil {
		return nil, error
	}
	return ngram.match(input, tval)
}

// BestMatch is the same as Search except that it's returning only one best result instead of all.
func (ngram *NGramIndex) BestMatch(input string, threshold ...float64) (*SearchResult, error) {
	if ngram.index == nil {
		ngram.init()
	}
	tval, error := validateThresholdValues(threshold)
	if error != nil {
		return nil, error
	}
	variants, error := ngram.match(input, tval)
	if error != nil {
		return nil, error
	}
	if len(variants) == 0 {
		return nil, errors.New("no matches found")
	}
	var result SearchResult
	maxsim := -1.0
	for _, val := range variants {
		if val.Similarity > maxsim {
			maxsim = val.Similarity
			result = val
		}
	}
	return &result, nil
}
