package clitools

import (
	"fmt"
	"github.com/urfave/cli"
)

// CheckRequired checks whether all required flags are set in the provided context.
func CheckRequired(c *cli.Context, required []string) error {
	for _, name := range required {
		if c.String(name) == "" {
			return fmt.Errorf(`"%s" is required`, name)
		}
	}

	return nil
}
