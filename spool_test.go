package ngram

import "testing"

func Test_spool(t *testing.T) {
	var spool stringPool
	ix0, err := spool.Append("foo")
	if err != nil {
		t.Error(err)
	}
	ix1, err := spool.Append("bar")
	if err != nil {
		t.Error(err)
	}
	st0, err := spool.ReadAt(ix0)
	if err != nil {
		t.Error(err)
	}
	if st0 != "foo" {
		t.Error("Can't read 'foo'")
	}
	st1, err := spool.ReadAt(ix1)
	if err != nil {
		t.Error(err)
	}
	if st1 != "bar" {
		t.Error("Can't read 'bar'")
	}
}
