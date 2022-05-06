#!/bin/bash
set -xe
# main.go is already excluded using https://pkg.go.dev/cmd/go#hdr-Build_constraints
CGO_ENABLED=1 go build -buildmode=c-shared -o ytarchive.so
python3 test_yta.py
python3 -c 'import ytarchive; print(ytarchive.sum(8,9))'
