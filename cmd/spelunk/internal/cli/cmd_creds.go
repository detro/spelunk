package cli

import (
	"context"
)

// CredsCmd detects and verifies the provided credentials for each known Secret Source.
type CredsCmd struct{}

func (c *CredsCmd) Run(cli *CLI) error {
	ctx := context.Background()
	return cli.Config.VerifyAll(ctx)
}
