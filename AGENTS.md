# AGENTS.md

## Project

`ssh-get-id` — single-binary Go CLI that fetches SSH public keys from online identity providers and merges them into `authorized_keys`.

- **Module**: `github.com/mengzhuo/ssh-get-id` (Go 1.25)
- **Deps**: only `golang.org/x/crypto` (SSH key parsing)
- **License**: MIT

## Architecture

- **`cmd/ssh-get-id/`**: CLI binary entry point — flag parsing, orchestration, file I/O.
- **Root package `sshgetid`**: reusable library — key parsing, merging, source fetching.
  - `keys.go`: `KeyTable`, `Entry`, `ParseKeys`, `MergeKeys`
  - `sources.go`: `Source` interface, `HTTPSource`, `SourceTable` (all 5 identity providers)
- Key dedup: by `ssh.MarshalAuthorizedKey` fingerprint; duplicates skipped with optional warning.
- Imported keys tagged: `#ssh-get-id <prefix>:<user>`.

## Commands

```bash
# Build
go build -o ssh-get-id ./cmd/ssh-get-id/
# or
make build

# Run unit tests
go test ./...
# or
make test

# Test with coverage
make cover

# Lint
make lint

# Run integration tests (require pre-built binary + network)
go build -o /usr/local/bin/ssh-get-id ./cmd/ssh-get-id/
bash ./tests/run.sh
bash ./tests/gh.sh
```

## Testing

- **Unit tests**: `keys_test.go`, `sources_test.go` — ParseKeys, MergeKeys, HTTPSource (with `httptest`), no network required.
- **Integration tests**: bash scripts in `tests/` that shell out to the compiled binary.
  - Tests hit the network (GitHub, Launchpad) — they fetch real SSH keys.
  - `tests/run.sh`: fetches keys to stdout and greps for the comment tag.
  - `tests/gh.sh`: writes to `~/.ssh/authorized_keys`, then verifies with grep.
- Run unit tests before integration: `go test ./... && go build ... && bash tests/*.sh`

## Linting

- Config: `.golangci.yml` (errcheck, staticcheck, govet, revive, misspell, etc.)
- CI runs `golangci/golangci-lint-action@v7` on every push/PR.

## Gotchas

- **`-k`**: skip TLS certificate verification (sets `InsecureSkipVerify` on the HTTP client). Use only for self-signed/internal CAs.
- **`-l NONE`** (uppercase): special value to skip loading local keys entirely. Different from empty string (which uses default path).
- **`-o -`**: output to stdout instead of a file.
- **Source prefixes are case-sensitive** — `gh:` works, `GH:` doesn't.

## Adding a new identity source

Add an entry to `SourceTable` in `sources.go`:

```go
var SourceTable = map[string]Source{
    // add here, e.g.:
    // "xx": HTTPSource("https://example.com/%s.keys"),
}
```

The `%s` in the URL template is replaced with the user ID.

## CI

- GitHub Actions: `.github/workflows/go.yml`
- On push/PR to `main`: lint → unit tests + coverage → Codecov upload → build → integration tests
- GoReleaser (`.goreleaser.yaml`): cross-compiles for linux/windows/darwin/freebsd/netbsd/openbsd with UPX compression
