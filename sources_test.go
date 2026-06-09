package sshgetid

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPSourceGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/testuser.keys" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGx_test_key user@test\n"))
	}))
	defer srv.Close()

	hs := HTTPSource(srv.URL + "/%s.keys")
	data, err := hs.Get("testuser")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "ssh-ed25519") {
		t.Errorf("unexpected data: %s", string(data))
	}
	if !strings.Contains(string(data), "user@test") {
		t.Errorf("expected user@test in data: %s", string(data))
	}
}

func TestHTTPSourceGetError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	hs := HTTPSource(srv.URL + "/%s.keys")
	_, err := hs.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should mention status code, got: %v", err)
	}
}

func TestSourceTableHasKnownProviders(t *testing.T) {
	providers := []string{"cb", "gh", "gl", "lp", "st"}
	for _, p := range providers {
		if _, ok := SourceTable[p]; !ok {
			t.Errorf("missing expected provider: %s", p)
		}
	}
}

func TestHTTPSourceCustomClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "true" {
			t.Error("custom transport was not used")
		}
		w.Write([]byte("ok\n"))
	}))
	defer srv.Close()

	orig := HTTPClient
	defer func() { HTTPClient = orig }()
	HTTPClient = &http.Client{
		Transport: &customTransport{},
	}

	hs := HTTPSource(srv.URL + "/%s.keys")
	_, err := hs.Get("x")
	if err != nil {
		t.Fatal(err)
	}
}

type customTransport struct{}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-Custom", "true")
	return http.DefaultTransport.RoundTrip(req)
}
