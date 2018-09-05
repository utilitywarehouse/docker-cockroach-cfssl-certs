package internal

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/cloudflare/cfssl/csr"
	"github.com/urfave/cli"
)

const unknownCertTypeErrorTemplate = `unknown certificate type "%s", allowed values are "client" and "node"`

func fetchAndSaveLocalCerts(c *cli.Context) ([]byte, error) {
	var req *csr.CertificateRequest
	var keyFileName, certFileName string

	switch certificateType := c.GlobalString("type"); certificateType {
	case "client":
		user := c.GlobalString("user")
		req = newClientCSR(user)
		keyFileName = fmt.Sprintf("client.%s.key", user)
		certFileName = fmt.Sprintf("client.%s.crt", user)
	case "node":
		req = newNodeCSR(c.GlobalStringSlice("host"))
		keyFileName = "node.key"
		certFileName = "node.crt"
	default:
		return nil, fmt.Errorf(unknownCertTypeErrorTemplate, certificateType)
	}

	key, cert, err := createCertificateAndKey(
		c.GlobalString("ca-address"),
		c.GlobalString("ca-profile"),
		c.GlobalString("ca-auth-key"),
		req,
	)
	if err != nil {
		return nil, err
	}

	certsDir := c.GlobalString("certs-dir")
	keyFileName = path.Join(certsDir, keyFileName)
	certFileName = path.Join(certsDir, certFileName)

	if err := ioutil.WriteFile(keyFileName, key, 0600); err != nil {
		return cert, err
	}
	return cert, ioutil.WriteFile(certFileName, cert, 0600)
}

// FetchAndSaveCerts is a command that fetches certificates from the CA and saves them to disk.
func FetchAndSaveCerts(c *cli.Context) error {
	caCert, err := getCACertificate(c.GlobalString("ca-address"))
	if err != nil {
		return err
	}

	caFileName := path.Join(c.GlobalString("certs-dir"), "ca.crt")
	if err = ioutil.WriteFile(caFileName, caCert, 0600); err != nil {
		return err
	}

	_, err = fetchAndSaveLocalCerts(c)
	return err
}
