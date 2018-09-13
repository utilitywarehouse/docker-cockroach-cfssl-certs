package global

import (
	"io/ioutil"
	"path"

	"github.com/urfave/cli"

	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/internal"
)

// FetchAndSaveCerts is a command that fetches certificates from the CA and saves them to disk.
func FetchAndSaveCerts(c *cli.Context) error {
	caCert, err := internal.GetCACertificate(c.GlobalString("ca-address"))
	if err != nil {
		return err
	}

	caFileName := path.Join(c.GlobalString("certs-dir"), "ca.crt")
	if err = ioutil.WriteFile(caFileName, caCert, 0600); err != nil {
		return err
	}

	_, err = internal.FetchAndSaveLocalCerts(c)
	return err
}
