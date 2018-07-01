// go-qrcode
// Copyright 2014 Tom Harwood

package qrcode

import (
	"strings"
	"testing"
)

func TestQRCodeMaxCapacity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestQRCodeCapacity")
	}

	tests := []struct {
		string         string
		numRepetitions int
	}{
		{
			"0",
			7089,
		},
		{
			"A",
			4296,
		},
		{
			"#",
			2953,
		},
		// Alternate byte/numeric data types. Optimises to 2,952 bytes.
		{
			"#1",
			1476,
		},
	}

	for _, test := range tests {
		_, err := New(strings.Repeat(test.string, test.numRepetitions), Low)

		if err != nil {
			t.Errorf("%d x '%s' got %s expected success", test.numRepetitions,
				test.string, err.Error())
		}
	}

	for _, test := range tests {
		_, err := New(strings.Repeat(test.string, test.numRepetitions+1), Low)

		if err == nil {
			t.Errorf("%d x '%s' chars encodable, expected not encodable",
				test.numRepetitions+1, test.string)
		}
	}
}

func TestQRCodeVersionCapacity(t *testing.T) {
	tests := []struct {
		version         int
		level           RecoveryLevel
		maxNumeric      int
		maxAlphanumeric int
		maxByte         int
	}{
		{
			1,
			Low,
			41,
			25,
			17,
		},
		{
			2,
			Low,
			77,
			47,
			32,
		},
		{
			2,
			Highest,
			34,
			20,
			14,
		},
		{
			40,
			Low,
			7089,
			4296,
			2953,
		},
		{
			40,
			Highest,
			3057,
			1852,
			1273,
		},
	}

	for i, test := range tests {
		numericData := strings.Repeat("1", test.maxNumeric)
		alphanumericData := strings.Repeat("A", test.maxAlphanumeric)
		byteData := strings.Repeat("#", test.maxByte)

		var n *QRCode
		var a *QRCode
		var b *QRCode
		var err error

		n, err = New(numericData, test.level)
		if err != nil {
			t.Fatal(err.Error())
		}

		a, err = New(alphanumericData, test.level)
		if err != nil {
			t.Fatal(err.Error())
		}

		b, err = New(byteData, test.level)
		if err != nil {
			t.Fatal(err.Error())
		}

		if n.VersionNumber != test.version {
			t.Fatalf("Test #%d numeric has version #%d, expected #%d", i,
				n.VersionNumber, test.version)
		}

		if a.VersionNumber != test.version {
			t.Fatalf("Test #%d alphanumeric has version #%d, expected #%d", i,
				a.VersionNumber, test.version)
		}

		if b.VersionNumber != test.version {
			t.Fatalf("Test #%d byte has version #%d, expected #%d", i,
				b.VersionNumber, test.version)
		}
	}
}

func TestQRCodeISOAnnexIExample(t *testing.T) {
	var q *QRCode
	q, err := New("01234567", Medium)

	if err != nil {
		t.Fatalf("Error producing ISO Annex I Example: %s, expected success",
			err.Error())
	}

	const expectedMask int = 2

	if q.mask != 2 {
		t.Errorf("ISO Annex I example mask got %d, expected %d\n", q.mask,
			expectedMask)
	}
}

func BenchmarkQRCodeURLSize(b *testing.B) {
	for n := 0; n < b.N; n++ {
		New("http://www.example.org", Medium)
	}
}

func BenchmarkQRCodeMaximumSize(b *testing.B) {
	for n := 0; n < b.N; n++ {
		// 7089 is the maximum encodable number of numeric digits.
		New(strings.Repeat("0", 7089), Low)
	}
}
