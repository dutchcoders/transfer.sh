package storage

import (
	"context"
	"testing"
)

func TestGetAwsConfigUsesStaticCredentialsWhenProvided(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "environment-access-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "environment-secret-key")
	t.Setenv("AWS_SESSION_TOKEN", "environment-session-token")

	cfg, err := getAwsConfig(context.Background(), "configured-access-key", "configured-secret-key")
	if err != nil {
		t.Fatalf("getAwsConfig returned an error: %v", err)
	}

	credentials, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		t.Fatalf("retrieve credentials: %v", err)
	}
	if credentials.AccessKeyID != "configured-access-key" || credentials.SecretAccessKey != "configured-secret-key" || credentials.SessionToken != "" {
		t.Fatalf("got credentials %q/%q, want configured static credentials", credentials.AccessKeyID, credentials.SecretAccessKey)
	}
}

func TestGetAwsConfigUsesDefaultCredentialChain(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "environment-access-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "environment-secret-key")
	t.Setenv("AWS_SESSION_TOKEN", "environment-session-token")

	cfg, err := getAwsConfig(context.Background(), "", "")
	if err != nil {
		t.Fatalf("getAwsConfig returned an error: %v", err)
	}

	credentials, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		t.Fatalf("retrieve credentials: %v", err)
	}
	if credentials.AccessKeyID != "environment-access-key" || credentials.SecretAccessKey != "environment-secret-key" || credentials.SessionToken != "environment-session-token" {
		t.Fatalf("got credentials %q/%q, want credentials from the default chain", credentials.AccessKeyID, credentials.SecretAccessKey)
	}
}

func TestGetAwsConfigRejectsPartialStaticCredentials(t *testing.T) {
	testCases := []struct {
		name      string
		accessKey string
		secretKey string
	}{
		{name: "missing secret key", accessKey: "access-key"},
		{name: "missing access key", secretKey: "secret-key"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := getAwsConfig(context.Background(), testCase.accessKey, testCase.secretKey); err == nil {
				t.Fatal("getAwsConfig returned no error for partial static credentials")
			}
		})
	}
}
