package main

import (
	"log"
	"os"
	"os/exec"
)

const golangCILintVersion = "v2.11.4"

func main() {
	cmd := exec.Command(
		"go", "run",
		"github.com/golangci/golangci-lint/v2/cmd/golangci-lint@"+golangCILintVersion,
		"run", "./...",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
