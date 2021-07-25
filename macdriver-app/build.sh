#!/bin/bash

# These env variables are needed to build the program on m1 macs
GOOS=darwin GOARCH=amd64 CGO_CFLAGS="-arch x86_64" CGO_ENABLED=1  go build .
