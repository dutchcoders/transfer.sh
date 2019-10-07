#!/bin/bash
set -xe

# Validate arguments
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <fuzz-type>"
    exit 1
fi

# Configure
NAME=transfersh
ROOT=./server
TYPE=$1

# Setup
export GOFUZZ111MODULE="on"
go get -u github.com/dvyukov/go-fuzz/go-fuzz github.com/dvyukov/go-fuzz/go-fuzz-build
go get -d -v -u ./...
if [ ! -f fuzzit ]; then
    wget -q -O fuzzit https://github.com/fuzzitdev/fuzzit/releases/download/v2.4.72/fuzzit_Linux_x86_64
    chmod a+x fuzzit
fi

# Fuzz
function fuzz {
    FUNC=Fuzz$1
    TARGET=$2
    DIR=${3:-$ROOT}
    go-fuzz-build -libfuzzer -func $FUNC -o fuzzer.a $DIR
    clang -fsanitize=fuzzer fuzzer.a -o fuzzer
    ./fuzzit create job --type $TYPE $NAME/$TARGET fuzzer
}
fuzz LocalStorage local-storage
