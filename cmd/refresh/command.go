package refresh

import (
	"time"

	"github.com/urfave/cli"

	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/cmd/global"
	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/pkg/clitools"
)

var Command = cli.Command{
	Name:   "refresh",
	Flags:  Flags,
	Action: clitools.CheckGlobalRequired(refreshCertificates, global.RequiredFlags),
	Usage:  "Periodically check and refresh certificates for cockroach node or client.",
}

var Flags = []cli.Flag{
	cli.IntFlag{
		Name:   "max-attempts",
		EnvVar: "MAX_ATTEMPTS",
		Value:  3,
		Usage:  "Maximum number of attempts to try to fetch new certificate.",
	},
	cli.DurationFlag{
		Name:   "extra-time",
		EnvVar: "EXTRA_TIME",
		Value:  time.Minute * 5,
		Usage: "Time by which we shorten the expiration to have time to get new certificates." +
			"This should be a positive duration.",
	},
	cli.StringFlag{
		Name:   "target-proc-command",
		EnvVar: "TARGET_PROC_COMMAND",
		Value:  "cockroach",
		Usage: "Substring of a command used to run the executable that" +
			"should be signalled to when a new certificate is retrieved.",
	},
	cli.StringFlag{
		Name:   "signal",
		EnvVar: "SIGNAL",
		Value:  "SIGHUP",
		Usage: "Signal to be send to the main process, when the certificates are refreshed. " +
			`Allowed values are "SIGHUP", "SIGTERM" and "SIGINT".`,
	},
	cli.DurationFlag{
		Name:   "max-random-sleep",
		EnvVar: "MAX_RANDOM_SLEEP",
		Value:  time.Duration(0),
		Usage: "Maximum random sleep time before sending the signal to the main process. " +
			"This is to prevent all containers being restarted at the same time.",
	},
}

// refreshCertificates is a command that periodically checks and refreshes certificates.
func refreshCertificates(c *cli.Context) error {
	refresher, err := NewRefresher(c)
	if err != nil {
		return err
	}

	return refresher.Run(c)
}
