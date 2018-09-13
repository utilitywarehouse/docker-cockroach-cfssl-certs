package refresh

import (
	"fmt"
	"github.com/mitchellh/go-ps"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/internal"
	"io/ioutil"
	"os"
	"path"
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
		return nil, fmt.Errorf(internal.UnknownCertTypeErrorTemplate, certificateType)
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

func getTargetProcess(c *cli.Context) (*os.Process, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, err
	}

	command := c.String("target-proc-command")
	for _, p := range processes {
		if p.Executable() == command {
			return os.FindProcess(p.Pid())
		}

	}
	return nil, errProcessNotFound
}
