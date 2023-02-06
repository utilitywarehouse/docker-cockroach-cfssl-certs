package main

import (
	_ "embed"
	"os"

	"github.com/cloudflare/cfssl/log"
	"github.com/urfave/cli"

	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/cmd/global"
	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/cmd/refresh"
	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/cmd/refreshandforward"
	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/pkg/clitools"
)

//go:generate sh get_version
//go:embed version.txt
var version string

func main() {
	app := cli.NewApp()
	app.Version = version
	app.Name = "cockroach-certs"

	// Default Command
	app.Flags = global.Flags
	app.Action = clitools.CheckGlobalRequired(global.FetchAndSaveCerts, global.RequiredFlags)
	app.Usage = "Fetch certificates for Cockroach DB from cfssl CA."

	app.Commands = []cli.Command{
		refresh.Command,
		refreshandforward.Command,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
