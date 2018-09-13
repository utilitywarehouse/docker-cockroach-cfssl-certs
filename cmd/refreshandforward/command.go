package refreshandforward

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudflare/cfssl/log"
	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/cmd/global"
	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/cmd/refresh"
	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/pkg/clitools"
	"github.com/utilitywarehouse/docker-cockroach-cfssl-certs/pkg/request"
)

var Command = cli.Command{
	Name:   "refresh-and-forward",
	Flags:  Flags,
	Action: clitools.CheckGlobalRequired(refreshAndForward, global.RequiredFlags),
	Usage: "Behaves the same way as refresh, but it also forwards " +
		"all requests to the provided host (the main container.)",
}

var Flags = append(
	refresh.Flags,
	cli.IntFlag{
		Name:   "http-port",
		EnvVar: "HTTP_PORT",
		Value:  8000,
		Usage:  "Port from to forward http requests.",
	},
	cli.StringFlag{
		Name:   "forward-host",
		EnvVar: "FORWARD_HOST",
		Value:  "localhost:8080",
		Usage:  "Host address to which all requests should be forwarded.",
	},
	cli.StringFlag{
		Name:   "forward-scheme",
		EnvVar: "FORWARD_SCHEME",
		Value:  "https",
		Usage:  `Scheme to use for forwarding the requests ("http" or "https").`,
	},
	cli.DurationFlag{
		Name:   "forward-timeout",
		EnvVar: "FORWARD_TIMEOUT",
		Value:  3 * time.Second,
		Usage:  `Timeout for request forward.`,
	},
)

func refreshAndForward(c *cli.Context) error {
	refresher, err := refresh.NewRefresher(c)
	if err != nil {
		return err
	}

	port := fmt.Sprintf(":%v", c.Int("http-port"))
	forwardScheme := c.String("forward-scheme")
	forwardHost := c.String("forward-host")

	if forwardScheme != "http" && forwardScheme != "https" {
		return errors.Errorf(`"%s" is not a valid forwardScheme, allowed values are "http" and "https"`, forwardScheme)
	}
	httpHandler := &request.Forwarder{
		HTTPClient: http.Client{
			Timeout: c.Duration("forward-timeout"),
		},
		ForwardScheme: forwardScheme,
		ForwardHost:   forwardHost,
	}

	if forwardScheme == "https" {
		httpHandler.HTTPClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // nolint
			},
		}
	}

	http.Handle("/", httpHandler)
	log.Infof("Forwarding requests from port %s to %s", port, forwardHost)

	errChan := make(chan error)

	go func() { errChan <- refresher.Run(c) }()
	go func() { errChan <- http.ListenAndServe(port, nil) }()

	return <-errChan
}
