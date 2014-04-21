package ngram

import (
	"bytes"
	"errors"
	"fmt"
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

func (pool *stringPool) Append(s string) (int, error) {
	fmt.Println("Append ", s)
	begin := pool.buffer.Len()
	bstr := []byte(s)
	bstr = smaz.Compress(bstr)
	n, error := pool.buffer.Write(bstr)
	if error != nil {
		return 0, error
	}
	end := pool.buffer.Len()
	fmt.Println("begin: ", begin, "end: ", end, "n: ", n)
	ixitem := len(pool.items)
	pool.items = append(pool.items, region{begin: begin, end: end})
	fmt.Println("buffer state: ", pool.buffer)
	fmt.Println("items state: ", pool.items)
	return ixitem, nil
}

func (pool *stringPool) ReadAt(index int) (string, error) {
	fmt.Println("ReadAt ", index)
	if index < 0 || index >= len(pool.items) {
		return "", errors.New("Index out of range")
	}
	item := pool.items[index]
	compressed := pool.buffer.Bytes()[item.begin:item.end]
	decompressed, error := smaz.Decompress(compressed)
	if error != nil {
		return "", error
	}
	return string(decompressed), nil
}
