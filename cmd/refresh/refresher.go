package refresh

import (
	"crypto/x509"
	"encoding/pem"
	"math/rand"
	"os"
	"syscall"
	"time"

	"github.com/cloudflare/cfssl/log"
	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/internal"
)

var errMaxSleepTimeTooBig = errors.New(`"max-sleep-time" must be smaller than "extra-time"`)
var errMaxNumberOfAttemptsReached = errors.New("reached maximum number of attempts")

// Refresher is an object that reads and refreshes an ssl certificate when run
type Refresher interface {
	Run(*cli.Context) error
}

type refresher struct {
	maxAttempts int

	extraTime   time.Duration // Duration to be subtracted from the  actual certificate expiration time
	certExpTime time.Time     // Actual certificate expiry time minus `extraTime`

	process *signalProcess
}

type signalProcess struct {
	maxSleepTime time.Duration // Maximum length of random sleep before sending signal
	signal       os.Signal
	commandName  string
}

// NewRefresher returns a new instance of a Refresher or an error
// if an instance can't be created from provided context
func NewRefresher(c *cli.Context) (Refresher, error) {
	conf := &refresher{
		maxAttempts: c.Int("max-attempts"),
		extraTime:   c.Duration("extra-time"),
	}

	if conf.maxAttempts <= 0 {
		return nil, errors.New(`"max-attempts" must be strictly larger than 0`)
	}

	cert, err := loadLocalCert(c)
	if err != nil {
		return nil, err
	}

	if conf.setCertExpTime(cert) != nil {
		return nil, err
	}

	if cmd := c.String("target-proc-command"); cmd != "" {
		target := &signalProcess{
			maxSleepTime: c.Duration("random-sleep"),
			commandName:  c.String("target-proc-command"),
		}

		if target.maxSleepTime > conf.extraTime {
			return nil, errMaxSleepTimeTooBig
		}

		switch sig := c.String("signal"); sig {
		case "SIGHUP":
			target.signal = syscall.SIGHUP
		case "SIGTERM":
			target.signal = syscall.SIGTERM
		case "SIGINT":
			target.signal = syscall.SIGINT
		default:
			return nil, errors.Errorf(`"%s" is not an allowed signal`, sig)
		}

		if _, err = getTargetProcess(target.commandName); err != nil {
			return nil, err
		}

		conf.process = target
	}

	return conf, nil
}

func (r *refresher) setCertExpTime(certData []byte) error {
	block, _ := pem.Decode(certData)
	if block == nil {
		return errors.Errorf("failed to parse provided certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	r.certExpTime = cert.NotAfter.Add(-r.extraTime)
	return nil
}

func (r *refresher) fetchCerts(c *cli.Context) error {
	for attempts := 0; attempts < r.maxAttempts; {
		cert, err := internal.FetchAndSaveLocalCerts(c)
		if err != nil {
			log.Error(errors.Wrap(err, "failed fetching certificate"))
			attempts++
			continue
		}

		if err := r.setCertExpTime(cert); err != nil {
			log.Error(errors.Wrap(err, "failed extracting exp time"))
			attempts++
			continue
		}
		return nil
	}
	return errMaxNumberOfAttemptsReached
}

func (r *refresher) sendSignal() error {
	if r.process == nil {
		log.Info("no process to signal, skipping")
		return nil
	}

	// Send signal to target process after random sleep
	if r.process.maxSleepTime != 0 {
		sleepTime := rand.Int63n(int64(r.process.maxSleepTime))
		time.Sleep(time.Duration(sleepTime))
	}

	process, err := getTargetProcess(r.process.commandName)
	if err != nil {
		return err
	}

	if err := process.Signal(r.process.signal); err != nil {
		return errors.Wrap(err, "failed sending signal")
	}
	return nil
}

// Run starts a process that periodically checks certificate validity and
// if it is close to expiring it requests a new one and notifies the main process
func (r *refresher) Run(c *cli.Context) error {
	for {
		if r.certExpTime.After(time.Now()) {
			sleepTime := r.certExpTime.Sub(time.Now())
			log.Infof("Cert Exp Time: %v, Sleeping for: %v", r.certExpTime, sleepTime)
			time.Sleep(sleepTime)
		} else {
			log.Info("Requesting a new certificate.")
			if err := r.fetchCerts(c); err != nil {
				return err
			}
			if err := r.sendSignal(); err != nil {
				return err
			}
		}
	}
}
