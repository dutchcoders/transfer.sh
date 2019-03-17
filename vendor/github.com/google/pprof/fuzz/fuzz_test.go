// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pprof

import (
	"io/ioutil"
	"runtime"
	"testing"

	"github.com/google/pprof/profile"
)

func TestParseData(t *testing.T) {
	if runtime.GOOS == "nacl" {
		t.Skip("no direct filesystem access on Nacl")
	}

	const path = "testdata/"
	files, err := ioutil.ReadDir(path)
	if err != nil {
		t.Errorf("Problem reading directory %s : %v", path, err)
	}
	for _, f := range files {
		file := path + f.Name()
		inbytes, err := ioutil.ReadFile(file)
		if err != nil {
			t.Errorf("Problem reading file: %s : %v", file, err)
			continue
		}
		profile.ParseData(inbytes)
	}
}
