package main

import (
	"bytes"
	"log"
	"os"

	"github.com/traefik/yaegi/extract"
)

func main() {
	ext := extract.Extractor{Dest: "scripting"}

	buf := bytes.Buffer{}
	_, err := ext.Extract("./scripting/s", "github.com/tomruk/kopyaship/scripting/s", &buf)
	if err != nil {
		log.Fatalln(err)
	}

	bytes := bytes.ReplaceAll(buf.Bytes(), []byte("Symbols["), []byte("symbols["))
	err = os.WriteFile("./scripting/symbols.go", bytes, 0644)
	if err != nil {
		log.Fatalln(err)
	}
}
