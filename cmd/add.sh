#!/usr/bin/env bash

cd /Users/tornikekeburia/go/src/github.com/tkeburia/argen && cobra add $1 && mv ./cmd/${1}.go /Users/tornikekeburia/sandbox/argen/cmd
