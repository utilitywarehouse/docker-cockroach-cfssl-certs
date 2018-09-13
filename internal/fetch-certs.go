package internal

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/cloudflare/cfssl/csr"
	"github.com/urfave/cli"
)

// UnknownCertTypeErrorTemplate is a string template for unknown certificate type error
const UnknownCertTypeErrorTemplate = `unknown certificate type "%s", allowed values are "client" and "node"`

// FetchAndSaveLocalCerts fetches and saves certificate for a node or a client.
func FetchAndSaveLocalCerts(c *cli.Context) ([]byte, error) {
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
		return nil, fmt.Errorf(UnknownCertTypeErrorTemplate, certificateType)
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
