package clitools

import (
	"fmt"
	"github.com/urfave/cli"
)

// CheckGlobalRequired returns a cli.ActionFunc that checks whether all global required flags
// are set in the provided context and then runs the provided function.
func CheckGlobalRequired(function cli.ActionFunc, required []string) cli.ActionFunc {
	return func(context *cli.Context) error {
		for _, name := range required {
			if context.GlobalString(name) == "" {
				return fmt.Errorf(`"%s" is required`, name)
			}
		}

		return function(context)
	}
}
