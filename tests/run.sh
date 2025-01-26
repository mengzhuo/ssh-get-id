#!/bin/bash

set -xe

ssh-get-id -l NONE -o - gh:mengzhuo | grep "#ssh-get-id gh:mengzhuo"
