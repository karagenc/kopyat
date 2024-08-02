package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/traefik/yaegi/extract"
)

func main() {
outer:
	for {
		files, err := os.ReadDir(".")
		if err != nil {
			log.Fatalln(err)
		}
		for _, file := range files {
			if file.Name() == "go.mod" {
				break outer
			}
		}
		err = os.Chdir("..")
		if err != nil {
			log.Fatalln(err)
		}
	}

	err := extractKopyatPkg()
	if err != nil {
		log.Fatalln(err)
	}
	err = extractGitHubPkgs(
		"github.com/mitchellh/go-homedir",
		"github.com/gen2brain/beeep",
	)
	if err != nil {
		log.Fatalln(err)
	}
}

func extractGitHubPkgs(gitHubPaths ...string) error {
	for _, gitHubPath := range gitHubPaths {
		err := extractGitHubPkg(gitHubPath)
		if err != nil {
			return fmt.Errorf("extracting %s: %v", gitHubPath, err)
		}
	}
	return nil
}

func extractGitHubPkg(gitHubPath string) error {
	var (
		pkgIdent, importPath = gitHubPath, gitHubPath
		ext                  = extract.Extractor{Dest: "symbols"}
		b                    = bytes.Buffer{}
		gitHubSuffix         = gitHubSuffix(importPath)
	)

	_, err := ext.Extract(pkgIdent, importPath, &b)
	if err != nil {
		return err
	}
	return os.WriteFile("./internal/scripting/symbols/symbols_"+gitHubSuffix+".go", b.Bytes(), 0644)
}

func extractKopyatPkg() error {
	var (
		ext = extract.Extractor{Dest: "symbols"}
		b   = bytes.Buffer{}
	)

	_, err := ext.Extract(".", "github.com/karagenc/kopyat", &b)
	if err != nil {
		return err
	}
	buf := b.Bytes()
	buf = bytes.Replace(buf, []byte("	\".\"\n"), nil, 1)
	return os.WriteFile("./internal/scripting/symbols/symbols_kopyat.go", buf, 0644)
}

func gitHubSuffix(path string) string {
	path = strings.TrimPrefix(path, "github.com/")
	return strings.ReplaceAll(path, "/", "_")
}
