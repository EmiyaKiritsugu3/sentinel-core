#!/bin/bash
go test -v ./internal/liveview/... -coverprofile=coverage.out
go tool cover -func=coverage.out
