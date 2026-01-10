package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	typeRe := regexp.MustCompile(`\btype\s+Web\b`)
	methodRe := regexp.MustCompile(`\bfunc\s+\(\s*[\w\d_]+\s+(\*?Web)\s*\)`)
	found := false
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if filepath.Ext(p) != ".go" {
			return nil
		}
		f, err := os.Open(p)
		if err != nil {
			return nil
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		lineNo := 0
		for scanner.Scan() {
			lineNo++
			line := scanner.Text()
			if typeRe.MatchString(line) {
				if !found {
					fmt.Println("Found declarations related to 'Web':")
					found = true
				}
				fmt.Printf("type Web: %s:%d\n", p, lineNo)
			}
			if methodRe.MatchString(line) {
				if !found {
					fmt.Println("Found declarations related to 'Web':")
					found = true
				}
				fmt.Printf("method (Web receiver): %s:%d\n", p, lineNo)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "walk error: %v\n", err)
		os.Exit(2)
	}
	if !found {
		fmt.Println("No 'Web' type or Web methods found in scanned files.")
	}
}
