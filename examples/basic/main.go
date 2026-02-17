package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/types"
)

func main() {
	var coordStr string
	flag.StringVar(&coordStr, "secret", "", "Secret coordinates (e.g. plain://my-secret)")
	flag.Parse()

	if coordStr == "" {
		log.Fatal("Please provide secret coordinates using -secret flag")
	}

	// 1. Parse the coordinates
	coord, err := types.NewSecretCoord(coordStr)
	if err != nil {
		log.Fatalf("Failed to parse coordinates: %v", err)
	}

	// 2. Initialize Spelunker
	// By default, it includes built-in sources (env, file, plain, base64)
	s := spelunk.NewSpelunker()

	// 3. Dig up the secret
	secret, err := s.DigUp(context.Background(), coord)
	if err != nil {
		log.Fatalf("Failed to dig up secret: %v", err)
	}

	fmt.Printf("Secret value: %s\n", secret)
}
