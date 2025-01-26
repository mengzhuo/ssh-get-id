#!/bin/bash

set -xe

ssh-get-id gh:mengzhuo
grep "#ssh-get-id gh:mengzhuo" ~/.ssh/authorized_keys
