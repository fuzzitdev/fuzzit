package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"path/filepath"
)

func Test_copy(t *testing.T) {
	_, err := copyFile(filepath.Join("../../testdata", "dummy_file.txt"), "/tmp/dst")
	if err != nil {
		t.Errorf("copyFile failed with %s", err)
	}

	want := "Success!"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(want))
	}))
	defer srv.Close()

	_, err = loadFromFile(filepath.Join("../../testdata/", "conf.json"))
	if err != nil {
		t.Errorf("failed loadFromFile %s", err)
	}

}

