package internal

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/cloudflare/cfssl/log"
	"github.com/mitchellh/go-ps"
	errors2 "github.com/pkg/errors"
	"github.com/urfave/cli"
)

var errProcessNotFound = errors.New("process not found")

func loadLocalCert(c *cli.Context) ([]byte, error) {
	var certFileName string

	switch certificateType := c.GlobalString("type"); certificateType {
	case "client":
		certFileName = fmt.Sprintf("client.%s.crt", c.GlobalString("user"))
	case "node":
		certFileName = "node.crt"
	default:
		return nil, fmt.Errorf(unknownCertTypeErrorTemplate, certificateType)
	}

	certsDir := c.GlobalString("certs-dir")
	certFileName = path.Join(certsDir, certFileName)

	file, err := os.Open(certFileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}

func getCertificateExpirationTime(c *cli.Context, certData []byte) (expTime time.Time, err error) {
	block, _ := pem.Decode(certData)
	if block == nil {
		return expTime, fmt.Errorf("failed to parse provided certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return expTime, err
	}

	return cert.NotAfter.Add(-c.Duration("extra-time")), nil
}

func getTargetProcess(c *cli.Context) (*os.Process, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, err
	}

	command := c.String("target-proc-command")
	for _, p := range processes {
		if strings.Contains(p.Executable(), command) {
			return os.FindProcess(p.Pid())
		}

	}
	return nil, errProcessNotFound
}

// RefreshCertificates is a command that periodically checks and refreshes certificates.
func RefreshCertificates(c *cli.Context) error {
	attempts, maxAttempts := 0, c.Int("max-attempts")
	if maxAttempts <= 0 {
		return errors.New(`"max-attempts" must be strictly larger than 0`)
	}

	cert, err := loadLocalCert(c)
	if err != nil {
		return err
	}
	expTime, err := getCertificateExpirationTime(c, cert)
	if err != nil {
		return err
	}

	targetProcess, err := getTargetProcess(c)
	if err != nil {
		return err
	}

	for attempts < maxAttempts {
		if expTime.After(time.Now()) {
			sleepTime := expTime.Sub(time.Now())
			log.Infof("Cert Exp Time: %v, Sleeping for: %v", expTime, sleepTime)
			time.Sleep(sleepTime)
		} else {
			log.Info("Requesting a new certificate.")
			cert, err = fetchAndSaveLocalCerts(c)
			if err != nil {
				attempts++
				continue
			}
			// Send signal to target process
			if err = targetProcess.Signal(syscall.SIGHUP); err != nil {
				attempts++
				continue
			}

			var newExpTime time.Time
			newExpTime, err = getCertificateExpirationTime(c, cert)
			if err != nil {
				attempts++
				continue
			}
			expTime = newExpTime
			attempts = 0
		}
	}
	return errors2.Wrap(err, "reached maximum number of attempts")
}
