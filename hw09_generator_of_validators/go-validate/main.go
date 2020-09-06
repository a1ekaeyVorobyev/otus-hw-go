package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Place your code here
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("Please enter the .go filename for go-validate")
	}
	nameFile := args[0]
	if len(args) == 2 && args[1] == "false" {
		geneateFileWithError = false
	}
	err := ParserFile(nameFile)
	fmt.Print(nameFile)
	if len(err) != 0 {
		vname := strings.ReplaceAll(nameFile, filepath.Ext(nameFile), "_validation_generated.go")
		log.Printf("All error write in file:%v", vname)
		os.Exit(1)
	}
	os.Exit(0)
}
