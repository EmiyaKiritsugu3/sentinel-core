#!/bin/bash
# Revert to a clean state
git checkout internal/liveview/api_test.go
# Delete everything after TestHandleGetStatus_DBError
sed -i '/func TestHandleGetStatus_DBError(t \*testing.T) {/,/^}/!b;//!d;/^}/!d' internal/liveview/api_test.go
git checkout HEAD -- internal/liveview/api_test.go
