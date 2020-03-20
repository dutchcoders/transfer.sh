// +build gofuzz

package server

import (
	"bytes"
	"io"
	"math/rand"
	"reflect"
)

const applicationOctetStream = "application/octet-stream"

// FuzzLocalStorage tests the Local Storage.
func FuzzLocalStorage(fuzz []byte) int {
	var fuzzLength = uint64(len(fuzz))
	if fuzzLength == 0 {
		return -1
	}

	storage, err := NewLocalStorage("/tmp", nil)
	if err != nil {
		panic("unable to create local storage")
	}

	token := Encode(10000000 + int64(rand.Intn(1000000000)))
	filename := Encode(10000000+int64(rand.Intn(1000000000))) + ".bin"

	input := bytes.NewReader(fuzz)
	err = storage.Put(token, filename, input, applicationOctetStream, fuzzLength)
	if err != nil {
		panic("unable to save file")
	}

	contentLength, err := storage.Head(token, filename)
	if err != nil {
		panic("not visible through head")
	}

	if contentLength != fuzzLength {
		panic("incorrect content length")
	}

	output, contentLength, err := storage.Get(token, filename)
	if err != nil {
		panic("not visible through get")
	}

	if contentLength != fuzzLength {
		panic("incorrect content length")
	}

	var length uint64
	b := make([]byte, len(fuzz))
	for {
		n, err := output.Read(b)
		length += uint64(n)
		if err == io.EOF {
			break
		}
	}

	if !reflect.DeepEqual(b, fuzz) {
		panic("incorrect content body")
	}

	if length != fuzzLength {
		panic("incorrect content length")
	}

	err = storage.Delete(token, filename)
	if err != nil {
		panic("unable to delete file")
	}

	_, err = storage.Head(token, filename)
	if !storage.IsNotExist(err) {
		panic("file not deleted")
	}

	return 1
}
