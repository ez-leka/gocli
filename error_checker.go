package gocli

import (
	"bytes"
	"errors"
	"text/template"
)

type ErrorChecker struct {
	flag      int
	Templates map[string]string
}

func (eh ErrorChecker) makeError(template_str string, obj interface{}) string {
	buf := bytes.NewBuffer(nil)

	templ := template.Must(template.New("err").Parse(template_str))
	templ.Execute(buf, obj)
	return buf.String()
}

func (eh ErrorChecker) Error(t string) error {
	return errors.New(eh.makeError(eh.Templates[t], eh))
}
