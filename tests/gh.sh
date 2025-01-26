#!/bin/bash

set -xe

ssh-get-id gh:mengzhuo
grep "#ssh-get-id gh:mengzhuo" ~/.ssh/authorized_keys

ssh-get-id gh:mengzhuo lp:mengzhuo1203
grep "#ssh-get-id gh:mengzhuo" ~/.ssh/authorized_keys
grep "#ssh-get-id lp:mengzhuo" ~/.ssh/authorized_keys
