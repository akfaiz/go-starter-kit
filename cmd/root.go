package cmd

import (
	"context"
	"log"

	"github.com/akfaiz/go-starter-kit/cmd/migrate"
	"github.com/akfaiz/go-starter-kit/cmd/serve"
	"github.com/urfave/cli/v3"
)

var cmd = &cli.Command{
	Name:  "go-vue-starter-kit",
	Usage: "A starter kit for building web applications with Go and Vue.js",
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
