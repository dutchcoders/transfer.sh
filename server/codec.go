/*
https://github.com/fs111/kurz.go/blob/master/src/codec.go

Originally written and Copyright (c) 2011 André Kelpe
Modifications Copyright (c) 2015 John Ko

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package server

import (
	"math"
	"math/rand"
	"strings"
)

const (
	// characters used for short-urls
	SYMBOLS = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// someone set us up the bomb !!
	BASE = float64(len(SYMBOLS))

	// init seed encode number
	INIT_SEED = float64(-1)
)

// encodes a number into our *base* representation
// TODO can this be made better with some bitshifting?
func Encode(number float64, length int64) string {
	if number == INIT_SEED {
		seed := math.Pow(float64(BASE), float64(length))
		number = seed + (rand.Float64() * seed) // start with seed to enforce desired length
	}

	rest := int64(math.Mod(number, BASE))
	// strings are a bit weird in go...
	result := string(SYMBOLS[rest])
	if rest > 0 && number-float64(rest) != 0 {
		newnumber := (number - float64(rest)) / BASE
		result = Encode(newnumber, length) + result
	} else {
		// it would always be 1 because of starting with seed and we want to skip
		return ""
	}

	return result
}

// Decodes a string given in our encoding and returns the decimal
// integer.
func Decode(input string) int64 {
	const floatbase = float64(BASE)
	l := len(input)
	var sum int = 0
	for index := l - 1; index > -1; index -= 1 {
		current := string(input[index])
		pos := strings.Index(SYMBOLS, current)
		sum = sum + (pos * int(math.Pow(floatbase, float64((l-index-1)))))
	}
	return int64(sum)
}
