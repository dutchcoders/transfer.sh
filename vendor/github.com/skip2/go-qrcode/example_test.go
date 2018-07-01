// go-qrcode
// Copyright 2014 Tom Harwood
/*
	Amendments Thu, 2017-December-14:
	- test integration (go test -v)
	- idiomatic go code
*/
package qrcode

import (
	"fmt"
	"os"
	"testing"
)

func TestExampleEncode(t *testing.T) {
	if png, err := Encode("https://example.org", Medium, 256); err != nil {
		t.Errorf("Error: %s", err.Error())
	} else {
		fmt.Printf("PNG is %d bytes long", len(png))
	}
}

func TestExampleWriteFile(t *testing.T) {
	filename := "example.png"
	if err := WriteFile("https://example.org", Medium, 256, filename); err != nil {
		if err = os.Remove(filename); err != nil {
			t.Errorf("Error: %s", err.Error())
		}
	}
}
