go-ngram
========

N-Gram index

## Key features

* GC friendly (all strings are pooled and compressed)
* Application agnostic (there is no notion of document or something that user needs to implement)
 

## Usage

```go
index := ngram.NewNgramIndex(3, '$')     // first argument is gram size
                                         // second - padding charcter
tokenId, err := index.Add("hello")       // tokenId is unique token Id. 
                                         // We can get original string using it
str, err := index.GetString(tokenId)     // str == _"hello world"_
tokenList, err := index.Search("world")  // tokenList is list of tokens with weights
```

