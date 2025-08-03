package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	commit := getGitCommit()
	date := time.Now().Format("02-01-2006")
	version := "v0.0.1"

	code := fmt.Sprintf(`package main

func init() {
	buildVersion = "%s"
	buildDate    = "%s"
	buildCommit  = "%s"
}
`, version, date, commit)

	err := os.WriteFile("build.go", []byte(code), 0644)
	if err != nil {
		log.Fatal("failed to write build info")
	}
}

func getGitCommit() string {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "N/A"
	}

	return strings.TrimSpace(string(out))
}
