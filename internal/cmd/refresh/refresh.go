package refresh

import (
	"time"

	"github.com/urfave/cli"
)

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
}
