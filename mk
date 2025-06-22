#!/bin/bash

go generate ./...
go build -v ./cmd/antgen
go build -v ./cmd/replay
go build -v -tags "sqlite_math_functions" ./cmd/tabula
go build -v ./cmd/convert
