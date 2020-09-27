#!/usr/bin/env bash

GOOS=js GOARCH=wasm go build -o starfield.wasm .

export GOPHERJS_GOROOT="$(go1.12.16 env GOROOT)"
gopherjs build -o starfield.js
