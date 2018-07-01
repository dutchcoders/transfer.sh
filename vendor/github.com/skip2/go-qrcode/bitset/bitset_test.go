// go-qrcode
// Copyright 2014 Tom Harwood

package bitset

import (
	rand "math/rand"
	"testing"
)

func TestNewBitset(t *testing.T) {
	tests := [][]bool{
		{},
		{b1},
		{b0},
		{b1, b0},
		{b1, b0, b1},
		{b0, b0, b1},
	}

	for _, v := range tests {
		result := New(v...)

		if !equal(result.Bits(), v) {
			t.Errorf("%s", result.String())
			t.Errorf("%v => %v, want %v", v, result.Bits(), v)
		}
	}
}

func TestAppend(t *testing.T) {
	randomBools := make([]bool, 128)

	rng := rand.New(rand.NewSource(1))

	for i := 0; i < len(randomBools); i++ {
		randomBools[i] = rng.Intn(2) == 1
	}

	for i := 0; i < len(randomBools)-1; i++ {
		a := New(randomBools[0:i]...)
		b := New(randomBools[i:]...)

		a.Append(b)

		if !equal(a.Bits(), randomBools) {
			t.Errorf("got %v, want %v", a.Bits(), randomBools)
		}
	}
}

func TestAppendByte(t *testing.T) {
	tests := []struct {
		initial  *Bitset
		value    byte
		numBits  int
		expected *Bitset
	}{
		{
			New(),
			0x01,
			1,
			New(b1),
		},
		{
			New(b1),
			0x01,
			1,
			New(b1, b1),
		},
		{
			New(b0),
			0x01,
			1,
			New(b0, b1),
		},
		{
			New(b1, b0, b1, b0, b1, b0, b1),
			0xAA, // 0b10101010
			2,
			New(b1, b0, b1, b0, b1, b0, b1, b1, b0),
		},
		{
			New(b1, b0, b1, b0, b1, b0, b1),
			0xAA, // 0b10101010
			8,
			New(b1, b0, b1, b0, b1, b0, b1, b1, b0, b1, b0, b1, b0, b1, b0),
		},
	}

	for _, test := range tests {
		test.initial.AppendByte(test.value, test.numBits)
		if !equal(test.initial.Bits(), test.expected.Bits()) {
			t.Errorf("Got %v, expected %v", test.initial.Bits(),
				test.expected.Bits())
		}
	}
}

func TestAppendUint32(t *testing.T) {
	tests := []struct {
		initial  *Bitset
		value    uint32
		numBits  int
		expected *Bitset
	}{
		{
			New(),
			0xAAAAAAAF,
			4,
			New(b1, b1, b1, b1),
		},
		{
			New(),
			0xFFFFFFFF,
			32,
			New(b1, b1, b1, b1, b1, b1, b1, b1, b1, b1, b1, b1, b1, b1, b1, b1, b1,
				b1, b1, b1, b1, b1, b1, b1, b1, b1, b1, b1, b1, b1, b1, b1),
		},
		{
			New(),
			0x0,
			32,
			New(b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0,
				b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0, b0),
		},
		{
			New(),
			0xAAAAAAAA,
			32,
			New(b1, b0, b1, b0, b1, b0, b1, b0, b1, b0, b1, b0, b1, b0, b1, b0, b1,
				b0, b1, b0, b1, b0, b1, b0, b1, b0, b1, b0, b1, b0, b1, b0),
		},
		{
			New(),
			0xAAAAAAAA,
			31,
			New(b0, b1, b0, b1, b0, b1, b0, b1, b0, b1, b0, b1, b0, b1, b0, b1,
				b0, b1, b0, b1, b0, b1, b0, b1, b0, b1, b0, b1, b0, b1, b0),
		},
	}

	for _, test := range tests {
		test.initial.AppendUint32(test.value, test.numBits)
		if !equal(test.initial.Bits(), test.expected.Bits()) {
			t.Errorf("Got %v, expected %v", test.initial.Bits(),
				test.expected.Bits())
		}
	}
}

func TestAppendBools(t *testing.T) {
	randomBools := make([]bool, 128)

	rng := rand.New(rand.NewSource(1))

	for i := 0; i < len(randomBools); i++ {
		randomBools[i] = rng.Intn(2) == 1
	}

	for i := 0; i < len(randomBools)-1; i++ {
		result := New(randomBools[0:i]...)
		result.AppendBools(randomBools[i:]...)

		if !equal(result.Bits(), randomBools) {
			t.Errorf("got %v, want %v", result.Bits(), randomBools)
		}
	}
}

func BenchmarkShortAppend(b *testing.B) {
	bitset := New()

	for i := 0; i < b.N; i++ {
		bitset.AppendBools(b0, b1, b0, b1, b0, b1, b0)
	}
}

func TestLen(t *testing.T) {
	randomBools := make([]bool, 128)

	rng := rand.New(rand.NewSource(1))

	for i := 0; i < len(randomBools); i++ {
		randomBools[i] = rng.Intn(2) == 1
	}

	for i := 0; i < len(randomBools)-1; i++ {
		result := New(randomBools[0:i]...)

		if result.Len() != i {
			t.Errorf("Len = %d, want %d", result.Len(), i)
		}
	}
}

func TestAt(t *testing.T) {
	test := []bool{b0, b1, b0, b1, b0, b1, b1, b0, b1}

	bitset := New(test...)
	for i, v := range test {
		result := bitset.At(i)

		if result != test[i] {
			t.Errorf("bitset[%d] => %t, want %t", i, result, v)
		}
	}
}

func equal(a []bool, b []bool) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestExample(t *testing.T) {
	b := New()                       // {}
	b.AppendBools(true, true, false) // {1, 1, 0}
	b.AppendBools(true)              // {1, 1, 0, 1}
	b.AppendByte(0x02, 4)            // {1, 1, 0, 1, 0, 0, 1, 0}

	expected := []bool{b1, b1, b0, b1, b0, b0, b1, b0}

	if !equal(b.Bits(), expected) {
		t.Errorf("Got %v, expected %v", b.Bits(), expected)
	}
}

func TestByteAt(t *testing.T) {
	data := []bool{b0, b1, b0, b1, b0, b1, b1, b0, b1}

	tests := []struct {
		index    int
		expected byte
	}{
		{
			0,
			0x56,
		},
		{
			1,
			0xad,
		},
		{
			2,
			0x2d,
		},
		{
			5,
			0x0d,
		},
		{
			8,
			0x01,
		},
	}

	for _, test := range tests {
		b := New()
		b.AppendBools(data...)

		result := b.ByteAt(test.index)

		if result != test.expected {
			t.Errorf("Got %#x, expected %#x", result, test.expected)
		}
	}
}

func TestSubstr(t *testing.T) {
	data := []bool{b0, b1, b0, b1, b0, b1, b1, b0}

	tests := []struct {
		start    int
		end      int
		expected []bool
	}{
		{
			0,
			8,
			[]bool{b0, b1, b0, b1, b0, b1, b1, b0},
		},
		{
			0,
			0,
			[]bool{},
		},
		{
			0,
			1,
			[]bool{b0},
		},
		{
			2,
			4,
			[]bool{b0, b1},
		},
	}

	for _, test := range tests {
		b := New()
		b.AppendBools(data...)

		result := b.Substr(test.start, test.end)

		expected := New()
		expected.AppendBools(test.expected...)

		if !result.Equals(expected) {
			t.Errorf("Got %s, expected %s", result.String(), expected.String())
		}
	}
}
