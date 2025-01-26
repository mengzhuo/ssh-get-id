#!/bin/bash

set -xe

go run ../main.go -l NONE -o - gh:mengzhuo | grep "#ssh-get-id gh:mengzhuo"
