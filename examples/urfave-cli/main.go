package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/types"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "spelunker",
		Usage: "Dig up secrets from various sources",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "secret",
				Aliases:  []string{"s"},
				Usage:    "Secret coordinates (e.g. plain://my-secret)",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			// 1. Parse the coordinates
			coordStr := c.String("secret")
			coord, err := types.NewSecretCoord(coordStr)
			if err != nil {
				return fmt.Errorf("failed to parse coordinates: %w", err)
			}

			// 2. Initialize Spelunker
			s := spelunk.NewSpelunker()

			// 3. Dig up the secret
			secret, err := s.DigUp(context.Background(), coord)
			if err != nil {
				return fmt.Errorf("failed to dig up secret: %w", err)
			}

			fmt.Printf("Secret value: %s\n", secret)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
