package main

import (
	"context"
	"fmt"
	"log"

	"github.com/alecthomas/kong"
	"github.com/detro/spelunk"
	"github.com/detro/spelunk/types"
)

// CLI struct using Kong for flag parsing
type CLI struct {
	// Kong automatically handles types that implement encoding.TextUnmarshaler,
	// which types.SecretCoord does.
	Secret types.SecretCoord `name:"secret" short:"s" help:"Secret coordinates (e.g. plain://my-secret)" required:""`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)

	// 1. Initialize Spelunker
	s := spelunk.NewSpelunker()

	// 2. Dig up the secret
	// We pass the parsed SecretCoord directly
	val, err := s.DigUp(context.Background(), &cli.Secret)
	if err != nil {
		log.Fatalf("Failed to dig up secret: %v", err)
	}

	fmt.Printf("Dug up secret: %s\n", val)
	ctx.Exit(0)
}
