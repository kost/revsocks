#!/bin/sh
# apk add go-cross

mkdir  release/
export GOX=$HOME/go/bin/gox
export CGO_ENABLED=0
export GOX_LINUX_AMD64_LDFLAGS="-extldflags -static -s -w"
export GOX_LINUX_386_LDFLAGS="-extldflags -static -s -w"

$GOX -ldflags="-s -w" -output="release/{{.Dir}}_{{.OS}}_{{.Arch}}" 2>release/error.log

