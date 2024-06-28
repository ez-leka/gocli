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
		Name:        "create",
		Description: "this is a description of `create` command",
		Usage: `
		Use create to create resources either from a supplied file or individually using availabel subcommands user and job.
		`,
		Category: &gocli.CommandCategory{
			Name:  "Beginner Commands",
			Order: 1,
		},
		Commands: []*gocli.Command{
			{
				Name: "user",
				Flags: []gocli.IFlag{
					&gocli.Flag[gocli.String]{
						Name:     "uname",
						Short:    'u',
						Usage:    "name of teh user to create",
						Required: true,
					},
				},
			},
			{
				Name: "job",
				Args: []gocli.IArg{
					&gocli.Arg[gocli.OneOf]{
						Name:     "job type",
						Usage:    "type of jon to create",
						Hints:    []string{"batch", "task"},
						Required: true,
					},
					&gocli.Arg[gocli.String]{
						Name:  "name",
						Usage: "name of job",
					},
					&gocli.Arg[gocli.String]{
						Name:  "access",
						Usage: "access permissions",
					},
				},
			},
		},
		Flags: []gocli.IFlag{&gocli.Flag[gocli.String]{
			Name:  "filename",
			Short: 'f',
			Usage: "Name of the file to do create from",
		},
		},
	})
	app.AddCommand(gocli.Command{
		Name:        "edit",
		Description: "this is description for `edit` command",
		Commands:    []*gocli.Command{},
		Category: &gocli.CommandCategory{
			Name:  "Beginner Commands",
			Order: 2,
		},
	})
	app.AddCommand(gocli.Command{
		Name:        "delete",
		Description: "this is description for `delete` command",
		Commands:    []*gocli.Command{},
		Category: &gocli.CommandCategory{
			Name:  "Intermediate Commands",
			Order: 1,
		},
		Args: []gocli.IArg{
			&gocli.Arg[gocli.OneOf]{
				Name:  "resource",
				Usage: "resource argument for specifying resorce",
				Hints: []string{"user", "job"},
			},
		},
	})

	app.AddFlags([]gocli.IFlag{
		&gocli.Flag[gocli.String]{
			Name:   "special",
			Usage:  "Special flag",
			Hidden: true,
		},
		&gocli.Flag[gocli.OneOf]{
			Name:  "name",
			Usage: "Name of the person flag",
			Hints: []string{"john", "mary", "jane"},
		},
		&gocli.Flag[gocli.String]{
			Name:   "server",
			Usage:  "Name of the server",
			Hidden: true,
		},
		&gocli.Flag[gocli.String]{
			Name:   "token",
			Usage:  "Access Token ",
			Hidden: true,
		},
	},
	)

	err := app.Run(os.Args)
	if err != nil {
		msg := err.Error()
		log.Fatalf("%s, try --help", msg)
	}

}
