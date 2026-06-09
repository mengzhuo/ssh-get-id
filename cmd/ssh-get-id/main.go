// Command ssh-get-id fetches SSH public keys from online identity providers
// and merges them into the local authorized_keys file.
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	sshgetid "github.com/mengzhuo/ssh-get-id"
)

var (
	output    = flag.String("o", "", "Destination of keys: default ~/.ssh/authorized_keys, - for stdout")
	noWarn    = flag.Bool("w", false, "Do not warn about imported keys")
	loadLocal = flag.String("l", "", "local keys path, default ~/.ssh/authorized_keys")
	insecure  = flag.Bool("k", false, "Skip TLS certificate verification")
)

func main() {
	flag.Parse()

	if *insecure {
		sshgetid.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}

	sshgetid.Warn = func(e *sshgetid.Entry) {
		log.Printf("Already authorized:%s", e)
	}

	remoteKeys, err := getRemoteKeys()
	if err != nil {
		log.Fatal(err)
	}

	localKeys, err := getLocalKeys()
	if err != nil {
		log.Fatal(err)
	}

	localKeys.MergeKeys(remoteKeys, !*noWarn)

	var target io.Writer

	if *output == "-" {
		target = os.Stdout
	} else {
		p := *output
		if p == "" {
			p = getDefaultSSHPath()
		}

		out, err := os.OpenFile(p, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Fatal(err)
		}
		target = out
		defer func() { _ = out.Sync() }()
	}

	for _, e := range localKeys.List {
		_, err = fmt.Fprintln(target, e)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getDefaultSSHPath() string {
	home, _ := os.UserHomeDir()
	sshDir := filepath.Join(home, ".ssh")
	sf := filepath.Join(home, ".ssh", "authorized_keys")
	stat, err := os.Stat(sshDir)
	if err != nil && os.IsNotExist(err) {
		_ = os.MkdirAll(filepath.Join(home, ".ssh"), 0700)
		return sf
	}

	if stat.IsDir() {
		return sf
	}
	return ""
}

func getLocalKeys() (*sshgetid.KeyTable, error) {
	kt := sshgetid.NewKeyTable()
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
		if os.IsNotExist(err) {
			return kt, nil
		}
		return nil, err
	}
	err = kt.ParseKeys(data)
	return kt, err
}

func getRemoteKeys() (*sshgetid.KeyTable, error) {
	kt := sshgetid.NewKeyTable()

	for _, arg := range flag.Args() {
		if !strings.Contains(arg, ":") {
			return nil, fmt.Errorf("invalid user format %q, expected PREFIX:USER (e.g. gh:octocat)", arg)
		}

		srcName, id, _ := strings.Cut(arg, ":")
		src, ok := sshgetid.SourceTable[srcName]
		if !ok {
			return nil, fmt.Errorf("unknown source %q, valid sources: %v", srcName, knownSources())
		}

		data, err := src.Get(id)
		if err != nil {
			return nil, fmt.Errorf("%s(%s):%v", srcName, id, err)
		}

		if len(data) < len("ssh-rsa") {
			continue
		}

		remoteKeys := sshgetid.NewKeyTable()
		err = remoteKeys.ParseKeys(data)
		if err != nil {
			return nil, err
		}
		for _, e := range remoteKeys.List {
			e.Comment = fmt.Sprintf("#ssh-get-id %s:%s", srcName, id)
		}
		kt.MergeKeys(remoteKeys, false)
	}

	return kt, nil
}

func knownSources() []string {
	var names []string
	for k := range sshgetid.SourceTable {
		names = append(names, k)
	}
	return names
}
