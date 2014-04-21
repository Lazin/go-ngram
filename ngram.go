package ngram

import (
	"errors"
	"fmt"
	"github.com/reusee/mmh3"
)

const maxN = 8

const defaultPad = "$"

const defaultN = 3

type nGramValue struct {
	items map[int]int
}

type NGramIndex struct {
	pad   string
	n     int
	spool stringPool
	index map[uint32]*nGramValue
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

func NewNGramIndex(n int, pad rune) (*NGramIndex, error) {
	if n < 2 || n > maxN {
		return nil, errors.New("Bad 'n' value for n-gram index")
	}
	ngram := new(NGramIndex)
	ngram.n = n
	ngram.pad = string(pad)
	ngram.init()
	return ngram, nil
}

func (ngram *NGramIndex) Add(input string) error {
	if ngram.index == nil {
		return errors.New("NGram index not initialized")
	}
	results, error := ngram.split_input(input)
	if error != nil {
		return error
	}
	for _, hash := range results {
		fmt.Println("hash value: ", hash)
		var val *nGramValue = nil
		var ok bool
		if val, ok = ngram.index[hash]; ok {
			fmt.Println("val found")
		} else {
			val = new(nGramValue)
			val.items = make(map[int]int)
			ngram.index[hash] = val
		}
		// insert string and counter
		ixstr, error := ngram.spool.Append(input)
		if error != nil {
			return error
		}
		if count, ok := val.items[ixstr]; ok {
			val.items[ixstr] = count + 1
		} else {
			val.items[ixstr] = 1
		}
	}
	return nil
}

func (ngram *NGramIndex) Search(input string) ([]string, error) {
	if ngram.index == nil {
		return nil, errors.New("NGram index not initialized")
	}
	results, error := ngram.split_input(input)
	if error != nil {
		return nil, error
	}
	output := make([]string, 0)
	for _, hash := range results {
		if val := ngram.index[hash]; val != nil {
			for key, _ := range val.items {
				item, error := ngram.spool.ReadAt(key)
				if error != nil {
					return nil, error
				}
				output = append(output, item)
			}
		}
	}
	return output, nil
}
