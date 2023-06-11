//go:build ignore

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ez-leka/gocli"
)

func check_err(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

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
	for key, _ := range gocli.GoCliStrings {
		explanation := "go tamplate or plain string"
		if strings.Contains(key, "Format") {
			explanation = "fmt.Printf style format string"
		}
		_, err = fmt.Fprintf(w, "\t\"%s\": `%s`,\n", key, explanation)
		check_err(err)
	}
	_, err = fmt.Fprintln(w, "}")
	check_err(err)
	err = w.Flush()
	check_err(err)
}
