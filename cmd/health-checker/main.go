package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/utilitywarehouse/docker-cockroach-tools/pkg/clitools"
)

var version = "replaced by `make static`"

var flags = []cli.Flag{
	cli.StringFlag{
		Name:   "log-level",
		Value:  "DEBUG",
		EnvVar: "LOG_LEVEL",
	},
	cli.StringFlag{
		Name:   "http-port",
		EnvVar: "HTTP_PORT",
		Usage:  "HTTP port which the server will listen on.",
	},
	cli.StringFlag{
		Name:   "certificate",
		EnvVar: "CERTIFICATE",
		Usage:  "Path to the certificate file for a Cockroach DB node.",
	},
	cli.DurationFlag{
		Name:   "cert-exp-diff",
		EnvVar: "CERT_EXP_DIFF",
		Value:  time.Duration(0),
		Usage:  "Duration that is subtracted from the certificate expiry. Thus making it expire earlier.",
	},
	cli.StringFlag{
		Name:   "forward-address",
		EnvVar: "FORWARD_ADDRESS",
		Usage:  `Address to which the incoming requests will be forwarded e.g. "localhost:8080".`,
	},
	cli.DurationFlag{
		Name:   "forward-timeout",
		EnvVar: "FORWARD_TIMEOUT",
		Value:  time.Second * 2,
		Usage:  "Duration that is subtracted from the certificate expiry. Thus making it expire earlier.",
	},
}

var requiredFlags = []string{
	"http-port",
	"certificate",
	"forward-address",
}

func setUpLogger(c *cli.Context) *log.Entry {
	ll, err := log.ParseLevel(c.String("log-level"))
	if err != nil {
		log.Fatal(err)
	}

	log.SetLevel(ll)
	log.SetFormatter(&log.JSONFormatter{})
	return log.WithField("version", version)
}

func getCertificateExpirationTime(certPath string) (*time.Time, error) {
	file, err := os.Open(certPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM in file: %s", certPath)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &cert.NotAfter, nil
}

func action(c *cli.Context) error {
	if err := clitools.CheckRequired(c, requiredFlags); err != nil {
		return err
	}

	certExpTime, err := getCertificateExpirationTime(c.String("certificate"))
	if err != nil {
		return err
	}

	expTime := certExpTime.Add(-c.Duration("cert-exp-diff"))
	logger := setUpLogger(c)

	httpHandler := &handler{
		client: http.Client{
			Timeout: c.Duration("forward-timeout"),
		},
		expTime: expTime,
		host:    c.String("forward-address"),
		logger:  logger,
	}

	port := fmt.Sprintf(":%v", c.Uint("http-port"))
	http.Handle("/health", httpHandler)
	logger.WithFields(log.Fields{
		"port":          port,
		"cert_exp_time": expTime,
	}).Debugln("HTTP server started.")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "cockroach-health"
	app.Usage = "Sidecar container that exposes health endpoint. " +
		"The endpoint checks certificate expiry and then forwards the request to the provided health endpoint."
	app.Version = version

	app.Flags = flags
	app.Action = action

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
