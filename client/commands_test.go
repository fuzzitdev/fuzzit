package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestDownloadAndExtractCorpus(t *testing.T) {
	unAuthenticatedClient, err := NewFuzzitClient("")
	if err != nil {
		t.Fatal(err)
	}
	unAuthenticatedClient.Org = "fuzzitdev"

	authenticatedClient, err := NewFuzzitClient(os.Getenv("FUZZIT_API_KEY"))
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		client *FuzzitClient
		target string
		err    string
	}{
		// Public Client
		{unAuthenticatedClient, "invalid-target", "401 Unauthorized"},
		{unAuthenticatedClient, "parse-complex", ""},
		{unAuthenticatedClient, "empty-corpus", "404 Not Found"},
		{unAuthenticatedClient, "auth-required", "401 Unauthorized"},

		{authenticatedClient, "invalid-target", "404 Not Found"},
		{authenticatedClient, "parse-complex", ""},
		{authenticatedClient, "empty-corpus", "404 Not Found"},
		{authenticatedClient, "auth-required", "404 Not Found"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("no-auth: target:%s, err:%s", tc.target, tc.err), func(t *testing.T) {
			dir, err := ioutil.TempDir("", "example")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)
			err = tc.client.DownloadAndExtractCorpus(dir, tc.target)
			if err != nil {
				if err.Error() != tc.err {
					t.Errorf("was expecting %s received %s", tc.err, err.Error())
				}
			} else {
				if tc.err != "" {
					t.Errorf("was excepting an error received none")
				}
			}
		})
	}
}

func TestCreateLocalJob(t *testing.T) {
	unAuthenticatedClient, err := NewFuzzitClient("")
	if err != nil {
		t.Fatal(err)
	}
	unAuthenticatedClient.Org = "fuzzitdev"

	authenticatedClient, err := NewFuzzitClient(os.Getenv("FUZZIT_API_KEY"))
	if err != nil {
		t.Fatal(err)
	}

	newJob := Job{}
	newJob.Host = "stretch-llvm8"

	testCases := []struct {
		client *FuzzitClient
		target string
		err    string
	}{
		{unAuthenticatedClient, "invalid-target", "fuzzer exited with 1"},
		{unAuthenticatedClient, "parse-complex", ""},
		{unAuthenticatedClient, "empty-corpus", ""},
		{unAuthenticatedClient, "auth-required", "fuzzer exited with 1"},

		{authenticatedClient, "invalid-target", ""},
		{authenticatedClient, "parse-complex", ""},
		{authenticatedClient, "empty-corpus", ""},
		{authenticatedClient, "auth-required", ""},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("target:%s, err:%s", tc.target, tc.err), func(t *testing.T) {
			newJob.Type = "regression"
			newJob.TargetId = tc.target
			newJob.Host = "gcr.io/fuzzit-public/stretch-llvm8:64bdedf"
			err := tc.client.CreateLocalJob(newJob, []string{"testdata/fuzzer.tar.gz"})
			if err != nil {
				if err.Error() != tc.err {
					t.Errorf("was expecting %s received %s", tc.err, err.Error())
				}
			} else {
				if tc.err != "" {
					t.Errorf("was excepting an error received none")
				}
			}
		})

	}
}
