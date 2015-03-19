go-ngram [![Build Status](https://travis-ci.org/Lazin/go-ngram.svg)](https://travis-ci.org/Lazin/go-ngram)
========

N-gram index for Go.

## Key features

* Unicode support.
* Append only. Data can't be deleted from index.
* GC friendly (all strings are pooled and compressed)
* Application agnostic (there is no notion of document or something that user needs to implement)
 

## Usage

```go
index := ngram.NewNgramIndex(ngram.SetN(3))
tokenId, err := index.Add("hello") 
str, err := index.GetString(tokenId)  // str == "hello"
resultsList, err := index.Search("world")
```

## TODO:

* Smoothing functions (Laplace etc)

[![GoDoc](https://godoc.org/github.com/Lazin/go-ngram?status.png)](https://godoc.org/github.com/Lazin/go-ngram)

[![docs examples](https://sourcegraph.com/api/repos/github.com/Lazin/go-ngram/.badges/docs-examples.png)](https://sourcegraph.com/github.com/Lazin/go-ngram)

[![library users](https://sourcegraph.com/api/repos/github.com/Lazin/go-ngram/.badges/library-users.png)](https://sourcegraph.com/github.com/Lazin/go-ngram)
