package main

import (
	"log"
	"os"

	"github.com/ez-leka/gocli"
)

func main() {
	app := gocli.New()
	app.ShellCompletion = true
	app.Description = `{{.Name}} is a test program for gocli`
	app.AddCommand(gocli.Command{
		Name: "create",
		Commands: []*gocli.Command{
			{
				Name: "user",
			},
			{
				Name: "job",
			},
		},
	})
	app.AddCommand(gocli.Command{
		Name:     "edit",
		Commands: []*gocli.Command{},
	})
	app.AddCommand(gocli.Command{
		Name:     "delete",
		Commands: []*gocli.Command{},
		Args: []gocli.IArg{
			&gocli.Arg[gocli.OneOf]{
				Name:  "resource",
				Hints: []string{"user", "job"},
			},
		},
	})
	app.AddFlags([]gocli.IFlag{
		&gocli.Flag[gocli.String]{
			Name: "filename",
		},
		&gocli.Flag[gocli.String]{
			Name: "flag",
		},
		&gocli.Flag[gocli.OneOf]{
			Name:  "name",
			Hints: []string{"john", "mary", "jane"},
		},
	},
	)
	err := app.Run(os.Args)
	if err != nil {
		msg := err.Error()
		log.Fatalf("%s, try --help", msg)
	}

}
