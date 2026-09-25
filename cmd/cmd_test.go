package cmd

import (
	"testing"

	"github.com/dutchcoders/transfer.sh/server/storage"
	"github.com/urfave/cli/v2"
)

func TestS3CredentialsTypeDefaultsToLegacy(t *testing.T) {
	for _, flag := range globalFlags {
		stringFlag, ok := flag.(*cli.StringFlag)
		if !ok || stringFlag.Name != "s3-credentials-type" {
			continue
		}
		if stringFlag.Value != storage.S3CredentialsTypeLegacy {
			t.Fatalf("s3-credentials-type default = %q, want %q", stringFlag.Value, storage.S3CredentialsTypeLegacy)
		}
		return
	}

	t.Fatal("s3-credentials-type flag not found")
}
