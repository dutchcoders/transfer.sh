// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package gosrc

import (
	"testing"
)

var lineCommentTests = []struct {
	in, out string
}{
	{"", ""},
	{"//line  1", "//       "},
	{"//line x\n//line y", "//      \n//      "},
	{"x\n//line ", "x\n//     "},
}

func TestOverwriteLineComments(t *testing.T) {
	for _, tt := range lineCommentTests {
		p := []byte(tt.in)
		OverwriteLineComments(p)
		s := string(p)
		if s != tt.out {
			t.Errorf("in=%q, actual=%q, expect=%q", tt.in, s, tt.out)
		}
	}
}
