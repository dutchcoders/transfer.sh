// go-qrcode
// Copyright 2014 Tom Harwood

package qrcode

import "testing"

func TestSymbolBasic(t *testing.T) {
	size := 10
	quietZoneSize := 4

	m := newSymbol(size, quietZoneSize)

	if m.size != size+quietZoneSize*2 {
		t.Errorf("Symbol size is %d, expected %d", m.size, size+quietZoneSize*2)
	}

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {

			v := m.get(i, j)

			if v != false {
				t.Errorf("New symbol not empty")
			}

			if !m.empty(i, j) {
				t.Errorf("New symbol is not empty")
			}

			value := i*j%2 == 0
			m.set(i, j, value)

			v = m.get(i, j)

			if v != value {
				t.Errorf("Symbol ignores set bits")
			}

			if m.empty(i, j) {
				t.Errorf("Symbol ignores set bits")
			}
		}
	}
}

func TestSymbolPenalties(t *testing.T) {
	tests := []struct {
		pattern          [][]bool
		expectedPenalty1 int
		expectedPenalty2 int
		expectedPenalty3 int
		expectedPenalty4 int
	}{
		{
			[][]bool{
				{b0, b1, b0, b1, b0, b1},
				{b1, b0, b1, b0, b1, b0},
				{b0, b1, b0, b1, b0, b1},
				{b1, b0, b1, b0, b1, b0},
				{b0, b1, b0, b1, b0, b1},
				{b1, b0, b1, b0, b1, b0},
			},
			0, // No adjacent modules of same color.
			0, // No 2x2+ sized blocks.
			0, // No 1:1:3:1:1 pattern.
			-1,
		},
		{
			[][]bool{
				{b0, b0, b0, b1, b0, b1},
				{b1, b0, b1, b0, b1, b0},
				{b0, b1, b0, b1, b0, b1},
				{b1, b0, b1, b0, b1, b0},
				{b0, b1, b0, b1, b0, b1},
				{b1, b0, b1, b0, b1, b0},
			},
			0, // 5 adjacent modules of same colour, score = 0.
			0, // No 2x2+ sized blocks.
			0, // No 1:1:3:1:1 pattern.
			-1,
		},
		{
			[][]bool{
				{b0, b0, b0, b0, b0, b0},
				{b1, b0, b1, b0, b1, b0},
				{b0, b1, b0, b1, b0, b1},
				{b1, b0, b1, b0, b1, b0},
				{b0, b1, b0, b1, b0, b1},
				{b1, b0, b1, b0, b1, b0},
			},
			4, // 6 adjacent modules of same colour, score = 3 + (6-5)
			0, // No 2x2+ sized blocks.
			0, // No 1:1:3:1:1 pattern.
			-1,
		},
		{
			[][]bool{
				{b0, b0, b0, b0, b0, b0, b0},
				{b1, b0, b1, b0, b1, b0, b1},
				{b1, b0, b0, b0, b0, b0, b1},
				{b1, b0, b1, b0, b1, b0, b1},
				{b1, b0, b0, b0, b0, b0, b1},
				{b1, b0, b1, b0, b1, b0, b1},
				{b1, b0, b0, b0, b0, b0, b0},
			},
			28, // 3+(7-5) + 3+(6-5) + 3+(6-5) + 3+(6-5) + 3+(7-5) + 3+(7-5) = 28
			0,  // No 2x2+ sized blocks.
			0,  // No 1:1:3:1:1 pattern.
			-1,
		},
		{
			[][]bool{
				{b0, b0, b0, b1, b0, b1},
				{b0, b0, b1, b0, b1, b0},
				{b0, b1, b0, b1, b0, b1},
				{b1, b0, b1, b1, b1, b0},
				{b0, b1, b1, b1, b0, b1},
				{b1, b0, b1, b0, b1, b0},
			},
			-1,
			6, // 3*(2-1)*(2-1) + 3(2-1)*(2-1)
			0, // No 1:1:3:1:1 pattern.
			-1,
		},
		{
			[][]bool{
				{b0, b0, b0, b0, b0, b1},
				{b0, b0, b0, b0, b0, b1},
				{b0, b0, b0, b0, b0, b1},
				{b0, b0, b0, b0, b0, b1},
				{b0, b0, b0, b0, b0, b1},
				{b0, b0, b0, b0, b0, b1},
			},
			-1,
			60, // 3 * (5-1) * (6-1)
			0,  // No 1:1:3:1:1 pattern.
			-1,
		},
		{
			[][]bool{
				{b0, b0, b0, b0, b0, b1},
				{b0, b0, b0, b0, b0, b1},
				{b1, b1, b0, b1, b0, b1},
				{b1, b1, b0, b1, b0, b1},
				{b1, b1, b0, b1, b0, b1},
				{b1, b1, b0, b1, b0, b1},
			},
			-1,
			21, // 3*(5-1)*(2-1) + 3*(2-1)*(4-1) = 3*4 + 3*3
			0,  // No 1:1:3:1:1 pattern.
			-1,
		},
		{
			[][]bool{
				{b0, b0, b0, b0, b1, b0, b1, b1, b1, b0, b1, b0},
				{b0, b0, b0, b0, b1, b0, b1, b1, b1, b0, b1, b0},
				{b0, b0, b0, b0, b1, b0, b1, b1, b1, b0, b1, b0},
				{b0, b0, b0, b0, b1, b0, b1, b1, b1, b0, b1, b0},
				{b0, b0, b0, b0, b1, b0, b1, b1, b1, b0, b1, b0},
				{b0, b0, b0, b0, b1, b0, b1, b1, b1, b0, b1, b0},
				{b0, b0, b0, b0, b1, b0, b1, b1, b1, b0, b1, b0},
				{b0, b0, b0, b0, b1, b0, b1, b1, b1, b0, b1, b0},
				{b0, b0, b0, b0, b1, b0, b1, b1, b1, b0, b1, b0},
				{b0, b0, b0, b0, b1, b0, b1, b1, b1, b0, b1, b0},
				{b0, b0, b0, b0, b1, b0, b1, b1, b1, b0, b1, b0},
				{b0, b0, b0, b0, b1, b0, b1, b1, b1, b0, b1, b0},
			},
			-1,
			-1,
			480, // 12* 1:1:3:1:1 patterns, 12 * 40.
			-1,
		},
		{
			[][]bool{
				{b1, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b1, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b1, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b1, b1, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b1, b1, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b1, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b1, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b1, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
			},
			-1,
			-1,
			80, // 2* 1:1:3:1:1 patterns, 2 * 40.
			-1,
		},
		{
			[][]bool{
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
			},
			-1,
			-1,
			-1,
			100, // 10 * (10 steps of 5% deviation from 50% black/white).
		},
		{
			[][]bool{
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
			},
			-1,
			-1,
			-1,
			100, // 10 * (10 steps of 5% deviation from 50% black/white).
		},
		{
			[][]bool{
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
			},
			-1,
			-1,
			-1,
			0, // Exactly 50%/50% black/white.
		},
		{
			[][]bool{
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
			},
			-1,
			-1,
			-1,
			20, // 10 * (2 steps of 5% deviation towards white).
		},
		{
			[][]bool{
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
			},
			-1,
			-1,
			-1,
			30, // 10 * (3 steps of 5% deviation towards white).
		},
		{
			[][]bool{
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b0},
				{b0, b0, b0, b0, b0, b0, b0, b0, b0, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
				{b1, b1, b1, b1, b1, b1, b1, b1, b1, b1},
			},
			-1,
			-1,
			-1,
			30, // 10 * (3 steps of 5% deviation towards white).
		},
	}

	for i, test := range tests {
		s := newSymbol(len(test.pattern[0]), 4)
		s.set2dPattern(0, 0, test.pattern)

		penalty1 := s.penalty1()
		penalty2 := s.penalty2()
		penalty3 := s.penalty3()
		penalty4 := s.penalty4()

		ok := true

		if test.expectedPenalty1 != -1 && test.expectedPenalty1 != penalty1 {
			ok = false
		}
		if test.expectedPenalty2 != -1 && test.expectedPenalty2 != penalty2 {
			ok = false
		}
		if test.expectedPenalty3 != -1 && test.expectedPenalty3 != penalty3 {
			ok = false
		}
		if test.expectedPenalty4 != -1 && test.expectedPenalty4 != penalty4 {
			ok = false
		}

		if !ok {
			t.Fatalf("Penalty test #%d p1=%d, p2=%d, p3=%d, p4=%d (expected p1=%d, p2=%d, p3=%d, p4=%d)", i, penalty1, penalty2, penalty3, penalty4,
				test.expectedPenalty1, test.expectedPenalty2, test.expectedPenalty3,
				test.expectedPenalty4)
		}
	}
}
