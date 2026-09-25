package cmd

import (
	"testing"

	"github.com/urfave/cli/v2"
)

func TestS3CredentialsTypeDefaultsToLegacy(t *testing.T) {
	for _, flag := range globalFlags {
		stringFlag, ok := flag.(*cli.StringFlag)
		if !ok || stringFlag.Name != "s3-credentials-type" {
			continue
		}
		if stringFlag.Value != s3CredentialsTypeLegacy {
			t.Fatalf("s3-credentials-type default = %q, want %q", stringFlag.Value, s3CredentialsTypeLegacy)
		}
		return
	}

	t.Fatal("s3-credentials-type flag not found")
}
