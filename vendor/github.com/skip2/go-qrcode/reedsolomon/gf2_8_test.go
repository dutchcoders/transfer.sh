// go-qrcode
// Copyright 2014 Tom Harwood

package reedsolomon

import "testing"

func TestGFMultiplicationIdentities(t *testing.T) {
	for i := 0; i < 256; i++ {
		value := gfElement(i)
		if gfMultiply(gfZero, value) != gfZero {
			t.Errorf("0 . %d != 0", value)
		}

		if gfMultiply(value, gfOne) != value {
			t.Errorf("%d . 1 == %d, want %d", value, gfMultiply(value, gfOne), value)
		}
	}
}

func TestGFMultiplicationAndDivision(t *testing.T) {
	// a * b == result
	var tests = []struct {
		a      gfElement
		b      gfElement
		result gfElement
	}{
		{0, 29, 0},
		{1, 1, 1},
		{1, 32, 32},
		{2, 4, 8},
		{16, 128, 232},
		{17, 17, 28},
		{27, 9, 195},
	}

	for _, test := range tests {
		result := gfMultiply(test.a, test.b)

		if result != test.result {
			t.Errorf("%d * %d = %d, want %d", test.a, test.b, result, test.result)
		}

		if test.b != gfZero && test.result != gfZero {
			b := gfDivide(test.result, test.a)

			if b != test.b {
				t.Errorf("%d / %d = %d, want %d", test.result, test.a, b, test.b)
			}
		}
	}
}

func TestGFInverse(t *testing.T) {
	for i := 1; i < 256; i++ {
		a := gfElement(i)
		inverse := gfInverse(a)

		result := gfMultiply(a, inverse)

		if result != gfOne {
			t.Errorf("%d * %d^-1 == %d, want %d", a, inverse, result, gfOne)
		}
	}
}

func TestGFDivide(t *testing.T) {
	for i := 1; i < 256; i++ {
		for j := 1; j < 256; j++ {
			// a * b == product
			a := gfElement(i)
			b := gfElement(j)
			product := gfMultiply(a, b)

			// product / b == a
			result := gfDivide(product, b)

			if result != a {
				t.Errorf("%d / %d == %d, want %d", product, b, result, a)
			}
		}
	}
}
