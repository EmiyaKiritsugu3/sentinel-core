#!/bin/bash
sed -i 's/"testing"/"os"\n\t"testing"/g' internal/liveview/api_test.go
