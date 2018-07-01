// go-qrcode
// Copyright 2014 Tom Harwood

package reedsolomon

import (
	"testing"
)

func TestGFPolyAdd(t *testing.T) {
	// a + b == result
	var tests = []struct {
		a      gfPoly
		b      gfPoly
		result gfPoly
	}{
		{
			gfPoly{[]gfElement{0, 0, 0}},
			gfPoly{[]gfElement{0}},
			gfPoly{[]gfElement{}},
		},
		{
			gfPoly{[]gfElement{1, 0}},
			gfPoly{[]gfElement{1, 0}},
			gfPoly{[]gfElement{0, 0}},
		},
		{
			gfPoly{[]gfElement{0xA0, 0x80, 0xFF, 0x00}},
			gfPoly{[]gfElement{0x0A, 0x82}},
			gfPoly{[]gfElement{0xAA, 0x02, 0xFF}},
		},
	}

	for _, test := range tests {
		result := gfPolyAdd(test.a, test.b)

		if !test.result.equals(result) {
			t.Errorf("%s * %s != %s (got %s)\n", test.a.string(false), test.b.string(false),
				test.result.string(false), result.string(false))
		}

		if len(result.term) > 0 && result.term[len(result.term)-1] == 0 {
			t.Errorf("Result's maximum term coefficient is zero")
		}
	}
}

func TestGFPolyequals(t *testing.T) {
	// a == b if isEqual
	var tests = []struct {
		a       gfPoly
		b       gfPoly
		isEqual bool
	}{
		{
			gfPoly{[]gfElement{0}},
			gfPoly{[]gfElement{0}},
			true,
		},
		{
			gfPoly{[]gfElement{1}},
			gfPoly{[]gfElement{0}},
			false,
		},
		{
			gfPoly{[]gfElement{1, 0, 1, 0, 1}},
			gfPoly{[]gfElement{1, 0, 1, 0, 1}},
			true,
		},
		{
			gfPoly{[]gfElement{1, 0, 1}},
			gfPoly{[]gfElement{1, 0, 1, 0, 0}},
			true,
		},
	}

	for _, test := range tests {
		isEqual := test.a.equals(test.b)

		if isEqual != test.isEqual {
			t.Errorf("%s and %s equality is %t (got %t)\n", test.a.string(false), test.b.string(false),
				test.isEqual, isEqual)
		}
	}
}

func TestGFPolyMultiply(t *testing.T) {
	// a * b == result
	var tests = []struct {
		a      gfPoly
		b      gfPoly
		result gfPoly
	}{
		{
			gfPoly{[]gfElement{0, 0, 1}},
			gfPoly{[]gfElement{9}},
			gfPoly{[]gfElement{0, 0, 9}},
		},
		{
			gfPoly{[]gfElement{0, 16, 1}},
			gfPoly{[]gfElement{128, 2}},
			gfPoly{[]gfElement{0, 232, 160, 2}},
		},
		{
			gfPoly{[]gfElement{254, 120, 88, 44, 11, 1}},
			gfPoly{[]gfElement{16, 2, 0, 51, 44}},
			gfPoly{[]gfElement{91, 50, 25, 184, 194, 105, 45, 244, 58, 44}},
		},
	}

	for _, test := range tests {
		result := gfPolyMultiply(test.a, test.b)

		if !test.result.equals(result) {
			t.Errorf("%s * %s = %s (got %s)\n",
				test.a.string(false),
				test.b.string(false),
				test.result.string(false),
				result.string(false))
		}
	}
}

func TestGFPolyRemainder(t *testing.T) {
	// numerator / denominator == quotient + remainder.
	var tests = []struct {
		numerator   gfPoly
		denominator gfPoly
		remainder   gfPoly
	}{
		{
			gfPoly{[]gfElement{1}},
			gfPoly{[]gfElement{1}},
			gfPoly{[]gfElement{0}},
		},
		{
			gfPoly{[]gfElement{1, 0}},
			gfPoly{[]gfElement{1}},
			gfPoly{[]gfElement{0}},
		},
		{
			gfPoly{[]gfElement{1}},
			gfPoly{[]gfElement{1, 0}},
			gfPoly{[]gfElement{1}},
		},
		{
			gfPoly{[]gfElement{1, 0, 1}},
			gfPoly{[]gfElement{0, 1}},
			gfPoly{[]gfElement{1}},
		},
		// (x^12 + x^10) / (x^10 + x^8 + x^5 + x^4 + x^2 + x^1 + x^0) =
		//  (x^10 + x^8 + x^5 + x^4 + x^2 + x^1 + x^0) * x^2 +
		//  (x^7 + x^6 + x^4 + x^3 + x^2) (the remainder)
		{
			gfPoly{[]gfElement{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1}},
			gfPoly{[]gfElement{1, 1, 1, 0, 1, 1, 0, 0, 1, 0, 1}},
			gfPoly{[]gfElement{0, 0, 1, 1, 1, 0, 1, 1}},
		},
		{
			gfPoly{[]gfElement{91, 50, 25, 184, 194, 105, 45, 244, 58, 44}},
			gfPoly{[]gfElement{254, 120, 88, 44, 11, 1}},
			gfPoly{[]gfElement{}},
		},
		{
			gfPoly{[]gfElement{0, 0, 0, 0, 0, 0, 195, 172, 24, 64}},
			gfPoly{[]gfElement{116, 147, 63, 198, 31, 1}},
			gfPoly{[]gfElement{48, 174, 34, 13, 134}},
		},
	}

	for _, test := range tests {
		remainder := gfPolyRemainder(test.numerator, test.denominator)

		if !test.remainder.equals(remainder) {
			t.Errorf("%s / %s, remainder = %s (got %s)\n",
				test.numerator.string(false),
				test.denominator.string(false),
				test.remainder.string(false),
				remainder.string(false))
		}
	}
}
