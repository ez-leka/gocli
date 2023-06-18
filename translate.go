//go:build ignore

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ez-leka/gocli"
)

func check_err(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	if len(os.Args) != 2 {
		fmt.Println(`
You must specify language

	Usage:
		go run translate.go <language>

	Example:
		go run translate.go french
	`)
		os.Exit(1)
	}
	language := os.Args[1]
	pkg := "translate"

	err := os.MkdirAll(pkg, 0755)
	check_err(err)

	file, err := filepath.Abs(filepath.Join(pkg, language) + ".go")
	check_err(err)

	f, err := os.Create(file)
	check_err(err)
	defer f.Close()

	w := bufio.NewWriter(f)

	_, err = fmt.Fprintf(w, "package %s\n\n", pkg)
	check_err(err)
	_, err = fmt.Fprintf(w, "import(\n\t\"github.com/ez-leka/gocli/i18n\"\n) \n\n")
	check_err(err)

	_, err = fmt.Fprintf(w, "var %sEntries = i18n.Entries{\n", language)
	check_err(err)
	for key, val := range gocli.GoCliStrings {
		_, err = fmt.Fprintf(w, "\t\"%s\": `%s`,\n", key, val)
		check_err(err)
	}
	_, err = fmt.Fprintln(w, "}")
	check_err(err)
	err = w.Flush()
	check_err(err)
}
