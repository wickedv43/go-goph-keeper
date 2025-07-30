package main

import (
	"bufio"
	"os"
	"strings"
)

func main() {
	input, _ := os.Open("coverage.tmp.out")
	defer input.Close()

	os.Remove("coverage.tmp.out")

	output, _ := os.Create("coverage.out")
	defer output.Close()

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "/mocks/") ||
			strings.Contains(line, "/internal/api/") ||
			strings.Contains(line, "/generate/") ||
			strings.Contains(line, "/cover/") {
			continue
		}
		output.WriteString(line + "\n")
	}
}
