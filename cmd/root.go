package cmd

import (
	"context"
	"log"

	"github.com/akfaiz/go-starter-kit/cmd/migrate"
	"github.com/akfaiz/go-starter-kit/cmd/serve"
	"github.com/urfave/cli/v3"
)

var cmd = &cli.Command{
	Name:  "go-starter-kit",
	Usage: "A starter kit for building Go applications",
	Commands: []*cli.Command{
		serve.Command,
		migrate.Command(),
	},
}

func Execute(args []string) {
	if err := cmd.Run(context.Background(), args); err != nil {
		log.Fatal(err)
	}
}
