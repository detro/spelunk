package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/types"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// AppConfig represents our application configuration
type AppConfig struct {
	// SecretCoord implements encoding.TextUnmarshaler, allowing
	// Viper (via mapstructure) to parse the string directly into the struct.
	APIToken types.SecretCoord `mapstructure:"api_token"`
}

func main() {
	// 1. Create a dummy config file for demonstration
	configFile := "config.yaml"
	configContent := []byte(`api_token: "plain://my-secret-api-token"`)
	if err := os.WriteFile(configFile, configContent, 0o644); err != nil {
		log.Fatal(err)
	}
	defer os.Remove(configFile) // Cleanup

	// 2. Configure Viper
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	var config AppConfig

	// 3. Unmarshal with DecodeHook
	// We need to tell Viper how to handle the types.SecretCoord (TextUnmarshaler).
	// We use mapstructure.StringToTextUnmarshalerHookFunc() to bridge string -> SecretCoord.
	opts := viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTextUnmarshalerHookFunc(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	))

	if err := viper.Unmarshal(&config, opts); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}

	// 4. Initialize Spelunker
	s := spelunk.NewSpelunker()

	// 5. Dig up the secret
	fmt.Printf("Resolving secret from coordinate: %s\n", config.APIToken.Location)
	secret, err := s.DigUp(context.Background(), &config.APIToken)
	if err != nil {
		log.Fatalf("Failed to dig up secret: %v", err)
	}

	fmt.Printf("Secret value: %s\n", secret)
}
