package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/detro/spelunk/cmd/spelunk/internal/logger"
	"github.com/detro/spelunk/v2"
	"github.com/detro/spelunk/v2/types"
)

type DigCmd struct {
	Coordinate string `arg:"" name:"coordinate" help:"Coordinates to the Secret to dig up."`
}

func (c *DigCmd) Run() error {
	ctx := context.Background()
	slog.Debug("Digging secret", "coord", c.Coordinate)

	coord, err := types.NewSecretCoord(c.Coordinate)
	if err != nil {
		slog.Error("Failed to parse secret coordinate", "err", err, "coord", c.Coordinate)
		return err
	}
	slog.Log(ctx, logger.LevelTrace, "Coordinate parsed", "coord", c.Coordinate, "parsed", coord)

	sp := spelunk.NewSpelunker()
	secret, err := sp.DigUp(context.Background(), coord)
	if err != nil {
		slog.Error("Failed to dig up secret", "err", err, "coord", c.Coordinate)
		return err
	}
	slog.Log(ctx, logger.LevelTrace, "Spelunker initialized")

	_, err = fmt.Fprint(os.Stdout, secret)
	if err != nil {
		slog.Error(
			"Failed to write secret to standard output",
			"err",
			err,
			"coord",
			c.Coordinate,
		)
	}
	return nil
}
