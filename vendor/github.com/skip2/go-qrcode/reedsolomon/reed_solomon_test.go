// go-qrcode
// Copyright 2014 Tom Harwood

package reedsolomon

import (
	"testing"

	bitset "github.com/skip2/go-qrcode/bitset"
)

func TestGeneratorPoly(t *testing.T) {
	var tests = []struct {
		degree    int
		generator gfPoly
	}{
		// x^2 + 3x^1 + 2x^0 (the shortest generator poly)
		{
			2,
			gfPoly{term: []gfElement{2, 3, 1}},
		},
		// x^5 + 31x^4 + 198x^3 + 63x^2 + 147x^1 + 116x^0
		{
			5,
			gfPoly{term: []gfElement{116, 147, 63, 198, 31, 1}},
		},
		// x^68 + 131x^67 + 115x^66 + 9x^65 + 39x^64 + 18x^63 + 182x^62 + 60x^61 +
		// 94x^60 + 223x^59 + 230x^58 + 157x^57 + 142x^56 + 119x^55 + 85x^54 +
		// 107x^53 + 34x^52 + 174x^51 + 167x^50 + 109x^49 + 20x^48 + 185x^47 +
		// 112x^46 + 145x^45 + 172x^44 + 224x^43 + 170x^42 + 182x^41 + 107x^40 +
		// 38x^39 + 107x^38 + 71x^37 + 246x^36 + 230x^35 + 225x^34 + 144x^33 +
		// 20x^32 + 14x^31 + 175x^30 + 226x^29 + 245x^28 + 20x^27 + 219x^26 +
		// 212x^25 + 51x^24 + 158x^23 + 88x^22 + 63x^21 + 36x^20 + 199x^19 + 4x^18 +
		// 80x^17 + 157x^16 + 211x^15 + 239x^14 + 255x^13 + 7x^12 + 119x^11 + 11x^10
		// + 235x^9 + 12x^8 + 34x^7 + 149x^6 + 204x^5 + 8x^4 + 32x^3 + 29x^2 + 99x^1
		// + 11x^0 (the longest generator poly)
		{
			68,
			gfPoly{term: []gfElement{11, 99, 29, 32, 8, 204, 149, 34, 12,
				235, 11, 119, 7, 255, 239, 211, 157, 80, 4, 199, 36, 63, 88, 158, 51, 212,
				219, 20, 245, 226, 175, 14, 20, 144, 225, 230, 246, 71, 107, 38, 107, 182,
				170, 224, 172, 145, 112, 185, 20, 109, 167, 174, 34, 107, 85, 119, 142,
				157, 230, 223, 94, 60, 182, 18, 39, 9, 115, 131, 1}},
		},
	}

	for _, test := range tests {
		generator := rsGeneratorPoly(test.degree)

		if !generator.equals(test.generator) {
			t.Errorf("degree=%d generator=%s, want %s", test.degree,
				generator.string(true), test.generator.string(true))
		}
	}
}

func TestEncode(t *testing.T) {
	var tests = []struct {
		numECBytes int
		data       string
		rsCode     string
	}{
		{
			5,
			"01000000 00011000 10101100 11000011 00000000",
			"01000000 00011000 10101100 11000011 00000000 10000110 00001101 00100010 10101110 00110000",
		},
		{
			10,
			"00010000 00100000 00001100 01010110 01100001 10000000 11101100 00010001 11101100 00010001 11101100 00010001 11101100 00010001 11101100 00010001",
			"00010000 00100000 00001100 01010110 01100001 10000000 11101100 00010001 11101100 00010001 11101100 00010001 11101100 00010001 11101100 00010001 10100101 00100100 11010100 11000001 11101101 00110110 11000111 10000111 00101100 01010101",
		},
	}

	for _, test := range tests {
		data := bitset.NewFromBase2String(test.data)
		rsCode := bitset.NewFromBase2String(test.rsCode)

		result := Encode(data, test.numECBytes)

		if !rsCode.Equals(result) {
			t.Errorf("data=%s, numECBytes=%d, encoded=%s, want %s",
				data.String(),
				test.numECBytes,
				result.String(),
				rsCode)
		}
	}
}
