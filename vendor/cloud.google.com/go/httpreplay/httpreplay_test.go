// Copyright 2018 Google LLC
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

// +build go1.8

package httpreplay_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/httpreplay"
	"cloud.google.com/go/internal/testutil"
	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

func TestIntegration_RecordAndReplay(t *testing.T) {
	httpreplay.DebugHeaders()
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}
	f, err := ioutil.TempFile("", "httpreplay")
	if err != nil {
		t.Fatal(err)
	}
	replayFilename := f.Name()
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(replayFilename)
	projectID := testutil.ProjID()
	if projectID == "" {
		t.Skip("Need project ID. See CONTRIBUTING.md for details.")
	}
	ctx := context.Background()

	// Record.
	initial := time.Now()
	ibytes, err := json.Marshal(initial)
	if err != nil {
		t.Fatal(err)
	}
	rec, err := httpreplay.NewRecorder(replayFilename, ibytes)
	if err != nil {
		t.Fatal(err)
	}
	hc, err := rec.Client(ctx, option.WithTokenSource(
		testutil.TokenSource(ctx, storage.ScopeFullControl)))
	if err != nil {
		t.Fatal(err)
	}
	wanta, wantc := run(t, hc)
	if err := rec.Close(); err != nil {
		t.Fatalf("rec.Close: %v", err)
	}

	// Replay.
	rep, err := httpreplay.NewReplayer(replayFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer rep.Close()
	hc, err = rep.Client(ctx)
	if err != nil {
		t.Fatal(err)
	}
	gota, gotc := run(t, hc)

	if diff := testutil.Diff(gota, wanta); diff != "" {
		t.Error(diff)
	}
	if !bytes.Equal(gotc, wantc) {
		t.Errorf("got %q, want %q", gotc, wantc)
	}
	var gotInitial time.Time
	if err := json.Unmarshal(rep.Initial(), &gotInitial); err != nil {
		t.Fatal(err)
	}
	if !gotInitial.Equal(initial) {
		t.Errorf("initial: got %v, want %v", gotInitial, initial)
	}
}

// TODO(jba): test errors

func run(t *testing.T, hc *http.Client) (*storage.BucketAttrs, []byte) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithHTTPClient(hc))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	b := client.Bucket(testutil.ProjID())
	attrs, err := b.Attrs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	obj := b.Object("replay-test")
	w := obj.NewWriter(ctx)
	if _, err := w.Write([]byte{150, 151, 152}); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	r, err := obj.NewReader(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return attrs, contents
}
