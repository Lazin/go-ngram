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

type stringPool struct {
	items  []region
	buffer bytes.Buffer
}

func (pool *stringPool) Append(s string) (TokenId, error) {
	begin := pool.buffer.Len()
	bstr := []byte(s)
	bstr = smaz.Compress(bstr)
	n, error := pool.buffer.Write(bstr)
	if error != nil {
		return 0, error
	}
	end := begin + n
	ixitem := TokenId(len(pool.items))
	pool.items = append(pool.items, region{begin: begin, end: end})
	return ixitem, nil
}

func (pool *stringPool) ReadAt(index TokenId) (string, error) {
	if index < TokenId(0) || index >= TokenId(len(pool.items)) {
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
