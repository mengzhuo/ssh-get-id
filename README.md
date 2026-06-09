ssh-get-id
===

[![Go Reference](https://pkg.go.dev/badge/github.com/mengzhuo/ssh-get-id.svg)](https://pkg.go.dev/github.com/mengzhuo/ssh-get-id)
[![Go Report Card](https://goreportcard.com/badge/github.com/mengzhuo/ssh-get-id)](https://goreportcard.com/report/github.com/mengzhuo/ssh-get-id)
[![codecov](https://codecov.io/gh/mengzhuo/ssh-get-id/branch/main/graph/badge.svg)](https://codecov.io/gh/mengzhuo/ssh-get-id)
[![CI](https://github.com/mengzhuo/ssh-get-id/actions/workflows/go.yml/badge.svg)](https://github.com/mengzhuo/ssh-get-id/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Same as ssh-import-id (also inspired by), but don't require ssh or python!

Works on Windows and Mac too!

Currently supported identities include Github, Gitlab, Launchpad, Codeberg and sourcehut.

Install
----

- Download from release page https://github.com/mengzhuo/ssh-get-id/releases
- build from source `go install github.com/mengzhuo/ssh-get-id@latest`

Usage
----

ssh-get-id uses short prefix to indicate the location of the online identity. For now, these are:

```
'cb:' for Codeberg
'gh:' for Github
'gl:' for Gitlab
'lp:' for Launchpad
'st:' for sourcehut
```
For example
```
ssh-get-id gh:mengzhuo
```

```
Usage of ssh-get-id [-h] [-o FILE] USERID [USERID ...]:
  -l string
        local keys path, default ~/.ssh/authorized_keys
  -o string
        Destination of keys: default ~/.ssh/authorized_keys, - for stdout
  -w    Do not warn about imported keys
  -k    Skip TLS certificate verification (insecure)
```

Build from source
```
go build -o ssh-get-id ./cmd/ssh-get-id/
```
