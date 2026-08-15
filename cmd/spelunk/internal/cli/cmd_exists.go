package cli

import (
	"context"
)

// ExistsCmd confirms existence of a secret at the given Coordinate.
type ExistsCmd struct {
	coordsArgs `embed:""`
}

func (c *ExistsCmd) Run(cli *CLI) error {
	ctx := context.Background()
	_, err := cli.DigUpSecret(ctx, c.Coordinate)
	if err != nil {
		return err
	}
	return nil
}
