package main

import (
	"bytes"
	"testing"
)

var testOkInput = `https://golang.org
/etc/passwd
https://golang.org
https://golang.org
`

var testOkResult = `Count for https://golang.org: 8
Count for /etc/passwd: 0
Count for https://golang.org: 8
Count for https://golang.org: 8
Total: 24
`

func TestOK(t *testing.T) {
	in := bytes.NewBufferString(testOkInput)
	out := bytes.NewBuffer(nil)
	getWordCountFromSources(in, out)
	result := out.String()
	if result != testOkResult {
		t.Errorf("Test OK failed, result not match")
	}
}

func BenchmarkFast(b *testing.B) {
	in := bytes.NewBufferString(testOkInput)
	out := bytes.NewBuffer(nil)

	for i := 0; i < b.N; i++ {
		getWordCountFromSources(in, out)
	}
}
