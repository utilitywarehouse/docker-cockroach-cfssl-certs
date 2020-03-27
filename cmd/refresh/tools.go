package refresh

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/cloudflare/cfssl/log"
	"github.com/mitchellh/go-ps"
	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/internal"
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

	return ioutil.ReadFile(certFileName)
}

func getTargetProcess(commandName string) (*os.Process, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, err
	}

	for _, p := range processes {
		if p.Executable() == commandName {
			log.Infof("%s does not match %s", p.Executable(), commandName)
			return os.FindProcess(p.Pid())
		}

	}
	return nil, errProcessNotFound
}
