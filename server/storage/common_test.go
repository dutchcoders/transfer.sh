package storage

import "testing"

func TestObjectPathJoinsWithSlashes(t *testing.T) {
	if got := ObjectPath("/base", "token", "file.txt"); got != "/base/token/file.txt" {
		t.Fatalf("ObjectPath = %q, want %q", got, "/base/token/file.txt")
	}
}
