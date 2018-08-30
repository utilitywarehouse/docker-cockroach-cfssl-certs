package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/log"
	"github.com/urfave/cli"
)

var version = "replaced by `make fast`"

var flags = []cli.Flag{
	cli.StringFlag{
		Name:   "type",
		Value:  "client",
		EnvVar: "CERTIFICATE_TYPE",
		Usage:  "Certificate type: \"node\" or \"client\"",
	},
	cli.StringFlag{
		Name:   "user",
		EnvVar: "USER",
		Usage:  "User name for client certificate",
	},
	cli.StringFlag{
		Name:  "hosts",
		Usage: "Coma separated list of host addresses for node certificate",
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

func checkRequired(c *cli.Context) error {
	if c.String("ca-auth-key") == "" {
		return errors.New("\"ca-auth-key\" is required")
	}
	if c.String("ca-profile") == "" {
		return errors.New("\"ca-profile\" is required")
	}
	if c.String("ca-address") == "" {
		return errors.New("\"ca-address\" is required")
	}

	return nil
}

func action(c *cli.Context) error {
	if err := checkRequired(c); err != nil {
		log.Fatal(err)
	}

	address := c.String("ca-address")
	caCert, err := getCACertificate(address)
	if err != nil {
		return err
	}

	var req *csr.CertificateRequest
	var keyFileName, certFileName string

	switch certificateType := c.String("type"); certificateType {
	case "client":
		user := c.String("user")
		req = newClientCSR(user)
		keyFileName = fmt.Sprintf("client.%s.key", user)
		certFileName = fmt.Sprintf("client.%s.crt", user)
	case "node":
		req = newNodeCSR(c.String("hosts"))
		keyFileName = "node.key"
		certFileName = "node.crt"
	default:
		return fmt.Errorf(
			"unknown certificate type \"%s\", allowed values are \"client\" and \"node\"", certificateType)
	}

	key, cert, err := createCertificateAndKey(address, c.String("ca-profile"), c.String("ca-auth-key"), req)
	if err != nil {
		return err
	}

	certsDir := c.String("certs-dir")
	caFileName := path.Join(certsDir, "ca.crt")
	keyFileName = path.Join(certsDir, keyFileName)
	certFileName = path.Join(certsDir, certFileName)

	if err := ioutil.WriteFile(caFileName, caCert, 0600); err != nil {
		return err
	}
	if err := ioutil.WriteFile(keyFileName, key, 0600); err != nil {
		return err
	}
	return ioutil.WriteFile(certFileName, cert, 0600)
}

func main() {
	app := cli.NewApp()
	app.Name = "cockroach-certs"
	app.Usage = "Fetch certificates for Cockroach DB from cfssl CA."
	app.Version = version

	app.Flags = flags
	app.Action = action

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
