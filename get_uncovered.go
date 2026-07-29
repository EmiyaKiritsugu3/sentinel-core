package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("coverage.out")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) >= 3 && strings.HasSuffix(parts[2], "0") {
			fmt.Println(line)
		}
	}
}
