ssh-get-id
===

Same as ssh-import-id (also inspired by), but don't require ssh or python!

Works on Windows and Mac too!

Currently supported identities include Github, Gitlab, Launchpad.

Usage
----

ssh-get-id uses short prefix to indicate the location of the online identity. For now, these are:

```
'cb:' for Codeberg
'gh:' for Github
'gl:' for Gitlab
'lp:' for Launchpad
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
```

Build from source
```
go build -o ssh-get-id main.go
```
