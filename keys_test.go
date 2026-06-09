package sshgetid

import (
	"crypto/ed25519"
	"crypto/rand"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func newTestKey(t *testing.T, comment string) ssh.PublicKey {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pub, err := ssh.NewPublicKey(priv.Public())
	if err != nil {
		t.Fatal(err)
	}
	return &commentKey{pub: pub, comment: comment}
}

type commentKey struct {
	pub     ssh.PublicKey
	comment string
}

func (k *commentKey) Type() string                              { return k.pub.Type() }
func (k *commentKey) Marshal() []byte                           { return k.pub.Marshal() }
func (k *commentKey) Verify(data []byte, sig *ssh.Signature) error { return k.pub.Verify(data, sig) }

func keyLine(t *testing.T, key ssh.PublicKey, comment string) string {
	t.Helper()
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key))) + " " + comment + "\n"
}

func TestNewKeyTable(t *testing.T) {
	kt := NewKeyTable()
	if kt == nil {
		t.Fatal("NewKeyTable returned nil")
	}
	if kt.m == nil {
		t.Fatal("internal map should be initialized")
	}
}

func TestParseKeys(t *testing.T) {
	key1 := newTestKey(t, "user@host1")
	key2 := newTestKey(t, "user@host2")

	data := []byte(keyLine(t, key1, "user@host1") + keyLine(t, key2, "user@host2"))

	kt := NewKeyTable()
	err := kt.ParseKeys(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(kt.List) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(kt.List))
	}
	if kt.List[0].Comment != "user@host1" {
		t.Errorf("unexpected comment: %q", kt.List[0].Comment)
	}
	if kt.List[1].Comment != "user@host2" {
		t.Errorf("unexpected comment: %q", kt.List[1].Comment)
	}
}

func TestParseKeysDedup(t *testing.T) {
	key := newTestKey(t, "dup@host")

	data := []byte(keyLine(t, key, "dup@host") + keyLine(t, key, "dup@host"))

	kt := NewKeyTable()
	err := kt.ParseKeys(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(kt.List) != 1 {
		t.Fatalf("expected 1 entry after dedup, got %d", len(kt.List))
	}
}

func TestParseKeysEmpty(t *testing.T) {
	kt := NewKeyTable()
	err := kt.ParseKeys([]byte{})
	if err != nil {
		t.Fatal(err)
	}
	if len(kt.List) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(kt.List))
	}
}

func TestParseKeysInvalid(t *testing.T) {
	kt := NewKeyTable()
	err := kt.ParseKeys([]byte("not a valid ssh key"))
	if err == nil {
		t.Fatal("expected error for invalid data")
	}
}

func TestMergeKeys(t *testing.T) {
	key1 := newTestKey(t, "key1")
	key2 := newTestKey(t, "key2")

	kt1 := NewKeyTable()
	data1 := []byte(keyLine(t, key1, "key1"))
	if err := kt1.ParseKeys(data1); err != nil {
		t.Fatal(err)
	}

	kt2 := NewKeyTable()
	data2 := []byte(keyLine(t, key2, "key2"))
	if err := kt2.ParseKeys(data2); err != nil {
		t.Fatal(err)
	}

	kt1.MergeKeys(kt2, false)

	if len(kt1.List) != 2 {
		t.Fatalf("expected 2 entries after merge, got %d", len(kt1.List))
	}
}

func TestMergeKeysDedup(t *testing.T) {
	key1 := newTestKey(t, "original")

	kt1 := NewKeyTable()
	if err := kt1.ParseKeys([]byte(keyLine(t, key1, "original"))); err != nil {
		t.Fatal(err)
	}

	kt2 := NewKeyTable()
	if err := kt2.ParseKeys([]byte(keyLine(t, key1, "duplicate"))); err != nil {
		t.Fatal(err)
	}

	kt1.MergeKeys(kt2, false)

	if len(kt1.List) != 1 {
		t.Fatalf("expected 1 entry after merge with duplicate, got %d", len(kt1.List))
	}
	if kt1.List[0].Comment != "original" {
		t.Errorf("expected original comment preserved, got %q", kt1.List[0].Comment)
	}
}

func TestMergeKeysWarn(t *testing.T) {
	key1 := newTestKey(t, "original")
	var warned *Entry

	Warn = func(e *Entry) { warned = e }
	defer func() { Warn = nil }()

	kt1 := NewKeyTable()
	if err := kt1.ParseKeys([]byte(keyLine(t, key1, "original"))); err != nil {
		t.Fatal(err)
	}

	kt2 := NewKeyTable()
	if err := kt2.ParseKeys([]byte(keyLine(t, key1, "duplicate"))); err != nil {
		t.Fatal(err)
	}

	kt1.MergeKeys(kt2, true)

	if warned == nil {
		t.Fatal("expected Warn callback to be called for duplicate")
	}
	if warned.Comment != "original" {
		t.Errorf("expected warning for original entry, got comment %q", warned.Comment)
	}
}

func TestMergeKeysWarnSuppressed(t *testing.T) {
	key1 := newTestKey(t, "original")
	warnCalled := false

	Warn = func(e *Entry) { warnCalled = true }
	defer func() { Warn = nil }()

	kt1 := NewKeyTable()
	if err := kt1.ParseKeys([]byte(keyLine(t, key1, "original"))); err != nil {
		t.Fatal(err)
	}

	kt2 := NewKeyTable()
	if err := kt2.ParseKeys([]byte(keyLine(t, key1, "duplicate"))); err != nil {
		t.Fatal(err)
	}

	kt1.MergeKeys(kt2, false)

	if warnCalled {
		t.Fatal("Warn should not be called when warn=false")
	}
}

func TestEntryString(t *testing.T) {
	key := newTestKey(t, "")
	e := &Entry{Key: key, Comment: "test-comment"}

	out := e.String()
	if !strings.Contains(out, "test-comment") {
		t.Errorf("String() should contain comment, got: %s", out)
	}
	if strings.HasSuffix(out, " ") {
		t.Errorf("String() should not have trailing space")
	}
}

func TestEntryStringNoComment(t *testing.T) {
	key := newTestKey(t, "")
	e := &Entry{Key: key}

	out := e.String()
	if strings.Contains(out, "  ") {
		t.Error("String() without comment should not have extra spaces")
	}
}

func TestParseKeysMultipleNewlines(t *testing.T) {
	key1 := newTestKey(t, "k1")
	key2 := newTestKey(t, "k2")

	data := []byte(keyLine(t, key1, "k1") + "\n\n" + keyLine(t, key2, "k2"))

	kt := NewKeyTable()
	err := kt.ParseKeys(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(kt.List) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(kt.List))
	}
}

func TestParseKeysTrailingNewline(t *testing.T) {
	key := newTestKey(t, "foo")
	data := []byte(keyLine(t, key, "foo") + "\n")

	kt := NewKeyTable()
	err := kt.ParseKeys(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(kt.List) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(kt.List))
	}
}

func TestParseKeysNilMap(t *testing.T) {
	kt := &KeyTable{}
	key := newTestKey(t, "test")
	err := kt.ParseKeys([]byte(keyLine(t, key, "test")))
	if err != nil {
		t.Fatal(err)
	}
	if len(kt.List) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(kt.List))
	}
}
