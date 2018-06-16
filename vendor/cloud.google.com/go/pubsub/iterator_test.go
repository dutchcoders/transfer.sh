// Copyright 2017 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pubsub

import (
	"testing"

	"cloud.google.com/go/internal/testutil"
)

func TestSplitRequestIDs(t *testing.T) {
	ids := []string{"aaaa", "bbbb", "cccc", "dddd", "eeee"}
	for _, test := range []struct {
		ids        []string
		splitIndex int
	}{
		{[]string{}, 0},
		{ids, 2},
		{ids[:2], 2},
	} {
		got1, got2 := splitRequestIDs(test.ids, reqFixedOverhead+20)
		want1, want2 := test.ids[:test.splitIndex], test.ids[test.splitIndex:]
		if !testutil.Equal(got1, want1) {
			t.Errorf("%v, 1: got %v, want %v", test, got1, want1)
		}
		if !testutil.Equal(got2, want2) {
			t.Errorf("%v, 2: got %v, want %v", test, got2, want2)
		}
	}
}
