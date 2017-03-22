// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package doc

type sliceWriter struct{ p *[]byte }

func (w sliceWriter) Write(p []byte) (int, error) {
	*w.p = append(*w.p, p...)
	return len(p), nil
}
