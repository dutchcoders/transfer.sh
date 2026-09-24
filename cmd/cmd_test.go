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

func TestResolveS3Credentials(t *testing.T) {
	testCases := []struct {
		name            string
		credentialsType string
		accessKey       string
		secretKey       string
		wantAccessKey   string
		wantSecretKey   string
		wantError       bool
	}{
		{
			name:            "legacy credentials",
			credentialsType: s3CredentialsTypeLegacy,
			accessKey:       "access-key",
			secretKey:       "secret-key",
			wantAccessKey:   "access-key",
			wantSecretKey:   "secret-key",
		},
		{
			name:            "legacy credentials missing access key",
			credentialsType: s3CredentialsTypeLegacy,
			secretKey:       "secret-key",
			wantError:       true,
		},
		{
			name:            "legacy credentials missing secret key",
			credentialsType: s3CredentialsTypeLegacy,
			accessKey:       "access-key",
			wantError:       true,
		},
		{
			name:            "default SDK credential chain",
			credentialsType: s3CredentialsTypeDefaultSDKCredentialChain,
			accessKey:       "ignored-access-key",
			secretKey:       "ignored-secret-key",
		},
		{
			name:            "unsupported credentials type",
			credentialsType: "unsupported",
			wantError:       true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			accessKey, secretKey, err := resolveS3Credentials(testCase.credentialsType, testCase.accessKey, testCase.secretKey)
			if (err != nil) != testCase.wantError {
				t.Fatalf("resolveS3Credentials() error = %v, wantError = %v", err, testCase.wantError)
			}
			if accessKey != testCase.wantAccessKey || secretKey != testCase.wantSecretKey {
				t.Fatalf("resolveS3Credentials() = %q/%q, want %q/%q", accessKey, secretKey, testCase.wantAccessKey, testCase.wantSecretKey)
			}
		})
	}
}
