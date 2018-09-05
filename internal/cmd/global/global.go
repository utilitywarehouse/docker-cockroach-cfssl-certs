package global

import "github.com/urfave/cli"

var Flags = []cli.Flag{
	cli.StringFlag{
		Name:   "type",
		Value:  "client",
		EnvVar: "CERTIFICATE_TYPE",
		Usage:  `Certificate type: "node" or "client"`,
	},
	cli.StringFlag{
		Name:   "user",
		EnvVar: "USER",
		Usage:  "User name for client certificate",
	},
	cli.StringSliceFlag{
		Name:  "host",
		Usage: `"--host=address1 --host=address2" One or more host addresses for the node certificate.`,
	},
	cli.StringFlag{
		Name:   "ca-auth-key",
		EnvVar: "CA_AUTH_KEY",
		Usage:  "Auth key to access the cfssl CA",
	},
	cli.StringFlag{
		Name:   "ca-profile",
		EnvVar: "CA_PROFILE",
		Usage:  "Profile to use when using cfssl CA",
	},
	cli.StringFlag{
		Name:   "ca-address",
		EnvVar: "CA_ADDRESS",
		Usage:  "Address of the cfssl CA",
	},
	cli.StringFlag{
		Name:   "certs-dir",
		Value:  "cockroach-certs",
		EnvVar: "CERTS_DIR",
		Usage:  "Directory where the certificates will be saved",
	},
}

var RequiredFlags = []string{
	"ca-auth-key",
	"ca-profile",
	"ca-address",
}
