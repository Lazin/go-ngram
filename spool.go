package ngram

import (
	"bytes"
	"errors"
	"github.com/cespare/go-smaz"
)

type region struct {
	begin int
	end   int
}

// string pool data structure
type stringPool struct {
	items  []region
	buffer bytes.Buffer
}

// Append adds new string to string pool. Function returns token ID and error.
// Strings doesn't need to be unique
func (pool *stringPool) Append(s string) (TokenID, error) {
	begin := pool.buffer.Len()
	bstr := []byte(s)
	bstr = smaz.Compress(bstr)
	n, error := pool.buffer.Write(bstr)
	if error != nil {
		return 0, error
	}
	end := begin + n
	ixitem := TokenID(len(pool.items))
	pool.items = append(pool.items, region{begin: begin, end: end})
	return ixitem, nil
}

// ReadAt converts token ID back to string.
func (pool *stringPool) ReadAt(index TokenID) (string, error) {
	if index < TokenID(0) || index >= TokenID(len(pool.items)) {
		return "", errors.New("Index out of range")
	}
	item := pool.items[int(index)]
	compressed := pool.buffer.Bytes()[item.begin:item.end]
	decompressed, error := smaz.Decompress(compressed)
	if error != nil {
		return "", error
	}
	return string(decompressed), nil
}
