#!/bin/bash
git checkout internal/liveview/api_test.go

# Find line of the first TestSetCORS_EmptyOrigin
line=$(grep -n "func TestSetCORS_EmptyOrigin(t \*testing.T) {" internal/liveview/api_test.go | head -n 2 | tail -n 1 | cut -d: -f1)

if [ ! -z "$line" ]; then
  # delete everything from the second occurrence to the end of the file
  sed -i "${line},\$d" internal/liveview/api_test.go
fi
