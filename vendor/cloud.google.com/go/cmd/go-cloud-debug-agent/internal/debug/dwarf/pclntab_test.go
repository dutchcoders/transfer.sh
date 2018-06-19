// Copyright 2018 Google Inc. All Rights Reserved.
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

package dwarf_test

// Stripped-down, simplified version of ../../gosym/pclntab_test.go

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	. "cloud.google.com/go/cmd/go-cloud-debug-agent/internal/debug/dwarf"
	"cloud.google.com/go/cmd/go-cloud-debug-agent/internal/debug/elf"
	"cloud.google.com/go/cmd/go-cloud-debug-agent/internal/debug/macho"
)

var (
	pclineTempDir    string
	pclinetestBinary string
)

func dotest(self bool) bool {
	// For now, only works on amd64 platforms.
	if runtime.GOARCH != "amd64" {
		return false
	}
	// Self test reads test binary; only works on Linux or Mac.
	if self {
		if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
			return false
		}
	}
	// Command below expects "sh", so Unix.
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		return false
	}
	if pclinetestBinary != "" {
		return true
	}
	var err error
	pclineTempDir, err = ioutil.TempDir("", "pclinetest")
	if err != nil {
		panic(err)
	}
	if strings.Contains(pclineTempDir, " ") {
		panic("unexpected space in tempdir")
	}
	// This command builds pclinetest from ../../gosym/pclinetest.asm;
	// the resulting binary looks like it was built from pclinetest.s,
	// but we have renamed it to keep it away from the go tool.
	pclinetestBinary = filepath.Join(pclineTempDir, "pclinetest")
	command := fmt.Sprintf("go tool asm -o %s.6 ../gosym/pclinetest.asm && go tool link -H %s -E main -o %s %s.6",
		pclinetestBinary, runtime.GOOS, pclinetestBinary, pclinetestBinary)
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	return true
}

func endtest() {
	if pclineTempDir != "" {
		os.RemoveAll(pclineTempDir)
		pclineTempDir = ""
		pclinetestBinary = ""
	}
}

func getData(file string) (*Data, error) {
	switch runtime.GOOS {
	case "linux":
		f, err := elf.Open(file)
		if err != nil {
			return nil, err
		}
		dwarf, err := f.DWARF()
		if err != nil {
			return nil, err
		}
		f.Close()
		return dwarf, nil
	case "darwin":
		f, err := macho.Open(file)
		if err != nil {
			return nil, err
		}
		dwarf, err := f.DWARF()
		if err != nil {
			return nil, err
		}
		f.Close()
		return dwarf, nil
	}
	panic("unimplemented DWARF for GOOS=" + runtime.GOOS)
}

func TestPCToLine(t *testing.T) {
	t.Skip("linker complains while building test binary")

	if !dotest(false) {
		return
	}
	defer endtest()

	data, err := getData(pclinetestBinary)
	if err != nil {
		t.Fatal(err)
	}

	// Test PCToLine.
	entry, err := data.LookupFunction("linefrompc")
	if err != nil {
		t.Fatal(err)
	}
	pc, ok := entry.Val(AttrLowpc).(uint64)
	if !ok {
		t.Fatal(`DWARF data for function "linefrompc" has no PC`)
	}
	for _, tt := range []struct {
		offset, want uint64
	}{
		{0, 2},
		{1, 3},
		{2, 4},
		{3, 4},
		{4, 5},
		{6, 5},
		{7, 6},
		{11, 6},
		{12, 7},
		{19, 7},
		{20, 8},
		{32, 8},
		{33, 9},
		{53, 9},
		{54, 10},
	} {
		file, line, err := data.PCToLine(pc + tt.offset)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasSuffix(file, "/pclinetest.asm") {
			t.Errorf("got %s; want %s", file, ".../pclinetest.asm")
		}
		if line != tt.want {
			t.Errorf("line for offset %d: got %d; want %d", tt.offset, line, tt.want)
		}
	}
}
