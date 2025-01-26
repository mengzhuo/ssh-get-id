package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

func getDefaultSSHPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ssh", "authorized_keys")
}

var (
	output = flag.String("o", "",
		"Destination of keys: default ~/.ssh/authorized_keys, - for stdout")
	noWarn    = flag.Bool("w", false, "Do not warn about imported keys")
	loadLocal = flag.String("l", "", "local keys path, default ~/.ssh/authorized_keys")
)

func getRemoteKeys() (kt *KeyTable, err error) {
	kt = newKt()

	for _, arg := range flag.Args() {

		if !strings.Contains(arg, ":") {
			flag.PrintDefaults()
			return
		}

		srcName, id, _ := strings.Cut(arg, ":")
		src, ok := SourceTable[srcName]
		if !ok {
			flag.PrintDefaults()
			return
		}

		data, err := src.Get(id)
		if err != nil {
			return nil, fmt.Errorf("%s(%s):%v", srcName, id, err)
		}

		if len(data) < len("ssh-rsa") {
			continue
		}

		remoteKeys := newKt()
		err = remoteKeys.parseKeys(data)
		if err != nil {
			return nil, err
		}
		for _, e := range remoteKeys.List {
			e.comment = fmt.Sprintf("#ssh-get-id %s:%s", srcName, id)
		}
		kt.mergeKeys(remoteKeys, false)
	}

	return kt, nil
}

func main() {

	flag.Parse()
	remoteKeys, err := getRemoteKeys()
	if err != nil {
		log.Fatal(err)
	}

	localKeys, err := getLocalKeys()
	if err != nil {
		log.Fatal(err)
	}

	localKeys.mergeKeys(remoteKeys, !*noWarn)

	var target io.Writer

	if *output == "-" {
		target = os.Stdout
	} else {
		out, err := os.OpenFile(*output, os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			log.Fatal(err)
		}
		target = out
		defer out.Close()
	}
	for _, e := range localKeys.List {
		_, err = fmt.Fprintln(target, e)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getLocalKeys() (*KeyTable, error) {

	kt := newKt()
	var path string
	switch *loadLocal {
	case "":
		path = getDefaultSSHPath()
	case "NONE":
		return kt, nil
	default:
		path = *loadLocal
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = kt.parseKeys(data)
	return kt, err
}

type Source interface {
	Get(id string) ([]byte, error)
}

var SourceTable = map[string]Source{
	"gh": httpSource("https://github.com/%s.keys"),
	"gl": httpSource("https://gitlab.com/%s.keys"),
	"lp": httpSource("https://launchpad.net/~%s/+sshkeys"),
}

type KeyTable struct {
	List []*Entry
	m    map[string]*Entry
}

type Entry struct {
	key     ssh.PublicKey
	comment string
	options []string
}

func (e *Entry) String() string {
	c := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(e.key)))
	if e.comment != "" {
		c += " " + e.comment
	}
	return c
}

func newKt() *KeyTable {
	return &KeyTable{m: make(map[string]*Entry)}
}

func (kt *KeyTable) parseKeys(data []byte) (err error) {

	if kt.m == nil {
		kt.m = make(map[string]*Entry)
	}
	for len(data) > 0 {
		key, comment, options, rest, err := ssh.ParseAuthorizedKey(data)
		if err != nil {
			return err
		}
		id := string(ssh.MarshalAuthorizedKey(key))
		if _, existed := kt.m[id]; !existed {
			e := &Entry{
				key:     key,
				comment: comment,
				options: options,
			}
			kt.m[id] = e
			kt.List = append(kt.List, e)
		}
		data = rest
	}
	return nil
}

func (kt *KeyTable) mergeKeys(rk *KeyTable, warn bool) {
	for hk, e := range rk.m {
		if ee, existed := kt.m[hk]; existed {
			if warn {
				log.Printf("Already authorized:%s", ee)
			}
			continue
		}
		kt.m[hk] = e
		kt.List = append(kt.List, e)
	}
}

type httpSource string

func (hs httpSource) Get(id string) (data []byte, err error) {
	gu, err := url.Parse(fmt.Sprintf(string(hs), id))
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(gu.String())
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}
