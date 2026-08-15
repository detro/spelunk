package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// DigCmd digs up a secret at the given Coordinate.
type DigCmd struct {
	coordsArgs `embed:""`
}

func (c *DigCmd) Run(cli *CLI) error {
	ctx := context.Background()
	secret, err := cli.DigUpSecret(ctx, c.Coordinate)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(os.Stdout, secret)
	if err != nil {
		slog.Error(
			"Failed to write secret to standard output",
			"err",
			err,
			"coord",
			c.Coordinate,
		)
		return err
	}
	return nil
}
