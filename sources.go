package sshgetid

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Source fetches SSH public keys for a given user ID.
type Source interface {
	Get(id string) ([]byte, error)
}

// HTTPClient is used by HTTPSource for all HTTP requests.
// Defaults to http.DefaultClient when nil. Set to a custom client
// to configure TLS verification, timeouts, or proxies.
var HTTPClient *http.Client

// HTTPSource is a URL template where %s is replaced with the user ID.
type HTTPSource string

// Get fetches keys from the URL template after substituting the user ID.
func (hs HTTPSource) Get(id string) ([]byte, error) {
	gu, err := url.Parse(fmt.Sprintf(string(hs), id))
	if err != nil {
		return nil, err
	}
	client := HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Get(gu.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned %d: %s", gu.Host, resp.StatusCode, body)
	}
	return body, nil
}

// SourceTable maps short prefixes to their Source implementation.
var SourceTable = map[string]Source{
	"cb": HTTPSource("https://codeberg.org/%s.keys"),
	"gh": HTTPSource("https://github.com/%s.keys"),
	"gl": HTTPSource("https://gitlab.com/%s.keys"),
	"lp": HTTPSource("https://launchpad.net/~%s/+sshkeys"),
	"st": HTTPSource("https://meta.sr.ht/~%s.keys"),
}
