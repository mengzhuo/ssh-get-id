// Package sshgetid provides SSH public key fetching, parsing, and merging.
package sshgetid

import (
	"bufio"
	"bytes"
	"strings"

	"golang.org/x/crypto/ssh"
)

// Entry represents a single authorized key entry with its parsed key, comment, and options.
type Entry struct {
	Key     ssh.PublicKey
	Comment string
	Options []string
}

// String returns the authorized_keys format line for this entry.
func (e *Entry) String() string {
	c := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(e.Key)))
	if e.Comment != "" {
		c += " " + e.Comment
	}
	return c
}

// KeyTable holds a deduplicated set of authorized key entries.
type KeyTable struct {
	List []*Entry
	m    map[string]*Entry
}

// NewKeyTable creates an empty KeyTable.
func NewKeyTable() *KeyTable {
	return &KeyTable{m: make(map[string]*Entry)}
}

// ParseKeys parses raw authorized_keys data into this KeyTable.
// Empty lines and leading whitespace are skipped. Duplicate keys (by fingerprint) are silently skipped.
func (kt *KeyTable) ParseKeys(data []byte) error {
	if kt.m == nil {
		kt.m = make(map[string]*Entry)
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		key, comment, options, _, err := ssh.ParseAuthorizedKey(line)
		if err != nil {
			return err
		}
		id := string(ssh.MarshalAuthorizedKey(key))
		if _, existed := kt.m[id]; !existed {
			e := &Entry{
				Key:     key,
				Comment: comment,
				Options: options,
			}
			kt.m[id] = e
			kt.List = append(kt.List, e)
		}
	}
	return scanner.Err()
}

// MergeKeys merges entries from rk into kt. Duplicates are skipped.
// If warn is true and Warn is non-nil, Warn is called for each duplicate.
func (kt *KeyTable) MergeKeys(rk *KeyTable, warn bool) {
	for hk, e := range rk.m {
		if ee, existed := kt.m[hk]; existed {
			if warn && Warn != nil {
				Warn(ee)
			}
			continue
		}
		kt.m[hk] = e
		kt.List = append(kt.List, e)
	}
}

// Warn is an optional callback invoked when MergeKeys finds a duplicate key.
// Set to nil (default) to suppress duplicate warnings.
var Warn func(entry *Entry)
