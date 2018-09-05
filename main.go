package main

import (
	"os"

	"github.com/cloudflare/cfssl/log"
	"github.com/urfave/cli"

	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/internal"
	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/internal/cmd/global"
	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/internal/cmd/refresh"
	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/pkg/clitools"
)

var version = "replaced by `make static`"

func main() {
	app := cli.NewApp()
	app.Version = version

	app.Name = "cockroach-certs"
	app.Flags = global.Flags
	app.Action = clitools.CheckGlobalRequired(internal.FetchAndSaveCerts, global.RequiredFlags)
	app.Usage = "Fetch certificates for Cockroach DB from cfssl CA."

	app.Commands = []cli.Command{
		{
			Name:   "refresh",
			Flags:  refresh.Flags,
			Action: clitools.CheckGlobalRequired(internal.RefreshCertificates, global.RequiredFlags),
			Usage:  "Periodically check and refresh certificates for cockroach node or client.",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
