#!/usr/bin/env bash

go build -v -ldflags="-X main.appVersion=$(git describe --all | cut -c7-32)"