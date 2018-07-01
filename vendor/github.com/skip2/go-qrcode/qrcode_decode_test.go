// go-qrcode
// Copyright 2014 Tom Harwood

package qrcode

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"testing"
)

// These tests use zbarimg to decode generated QR Codes to ensure they are
// readable. sudo apt-get install zbar-tools, or download from
// http://zbar.sourceforge.net.
//
// By default these tests are disabled to avoid a dependency on zbarimg if
// you're not running the tests. Use the -test-decode flag (go test
// -test-decode) to enable.

var testDecode *bool = flag.Bool("test-decode",
	false,
	"Enable decode tests. Requires zbarimg installed.")

var testDecodeFuzz *bool = flag.Bool("test-decode-fuzz",
	false,
	"Enable decode fuzz tests. Requires zbarimg installed.")

func TestDecodeBasic(t *testing.T) {
	if !*testDecode {
		t.Skip("Decode tests not enabled")
	}

	tests := []struct {
		content        string
		numRepetitions int
		level          RecoveryLevel
	}{
		{
			"A",
			1,
			Low,
		},
		{
			"A",
			1,
			Medium,
		},
		{
			"A",
			1,
			High,
		},
		{
			"A",
			1,
			Highest,
		},
		{
			"01234567",
			1,
			Medium,
		},
	}

	for _, test := range tests {
		content := strings.Repeat(test.content, test.numRepetitions)

		q, err := New(content, test.level)
		if err != nil {
			t.Error(err.Error())
		}

		err = zbarimgCheck(q)

		if err != nil {
			t.Error(err.Error())
		}
	}
}

func TestDecodeAllVersionLevels(t *testing.T) {
	if !*testDecode {
		t.Skip("Decode tests not enabled")
	}

	for version := 1; version <= 40; version++ {
		for _, level := range []RecoveryLevel{Low, Medium, High, Highest} {
			t.Logf("Version=%d Level=%d",
				version,
				level)

			q, err := newWithForcedVersion(
				fmt.Sprintf("v-%d l-%d", version, level), version, level)
			if err != nil {
				t.Fatal(err.Error())
				return
			}

			err = zbarimgCheck(q)

			if err != nil {
				t.Errorf("Version=%d Level=%d, err=%s, expected success",
					version,
					level,
					err.Error())
				continue
			}
		}
	}
}

func TestDecodeAllCharacters(t *testing.T) {
	if !*testDecode {
		t.Skip("Decode tests not enabled")
	}

	var content string

	// zbarimg has trouble with null bytes, hence start from ASCII 1.
	for i := 1; i < 256; i++ {
		content += string(i)
	}

	q, err := New(content, Low)
	if err != nil {
		t.Error(err.Error())
	}

	err = zbarimgCheck(q)

	if err != nil {
		t.Error(err.Error())
	}
}

func TestDecodeFuzz(t *testing.T) {
	if !*testDecodeFuzz {
		t.Skip("Decode fuzz tests not enabled")
	}

	r := rand.New(rand.NewSource(0))

	const iterations int = 32
	const maxLength int = 128

	for i := 0; i < iterations; i++ {
		len := r.Intn(maxLength-1) + 1

		var content string
		for j := 0; j < len; j++ {
			// zbarimg seems to have trouble with special characters, test printable
			// characters only for now.
			content += string(32 + r.Intn(94))
		}

		for _, level := range []RecoveryLevel{Low, Medium, High, Highest} {
			q, err := New(content, level)
			if err != nil {
				t.Error(err.Error())
			}

			err = zbarimgCheck(q)

			if err != nil {
				t.Error(err.Error())
			}
		}
	}
}

func zbarimgCheck(q *QRCode) error {
	s, err := zbarimgDecode(q)
	if err != nil {
		return err
	}

	if s != q.Content {
		q.WriteFile(256, fmt.Sprintf("%x.png", q.Content))
		return fmt.Errorf("got '%s' (%x) expected '%s' (%x)", s, s, q.Content, q.Content)
	}

	return nil
}

func zbarimgDecode(q *QRCode) (string, error) {
	var png []byte

	// 512x512px
	png, err := q.PNG(512)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("zbarimg", "--quiet", "-Sdisable",
		"-Sqrcode.enable", "/dev/stdin")

	var out bytes.Buffer

	cmd.Stdin = bytes.NewBuffer(png)
	cmd.Stdout = &out

	err = cmd.Run()

	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(strings.TrimPrefix(out.String(), "QR-Code:"), "\n"), nil
}

func BenchmarkDecodeTest(b *testing.B) {
	if !*testDecode {
		b.Skip("Decode benchmarks not enabled")
	}

	for n := 0; n < b.N; n++ {
		q, err := New("content", Medium)
		if err != nil {
			b.Error(err.Error())
		}

		err = zbarimgCheck(q)

		if err != nil {
			b.Error(err.Error())
		}
	}
}
