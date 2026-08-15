package cli

import (
	"context"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/detro/spelunk/cmd/spelunk/internal/logger"
	"github.com/detro/spelunk/plugin/modifier/jsonpath/v2"
	"github.com/detro/spelunk/plugin/modifier/tomlpath/v2"
	"github.com/detro/spelunk/plugin/modifier/xpath/v2"
	"github.com/detro/spelunk/plugin/modifier/yamlpath/v2"
	"github.com/detro/spelunk/v2"
	"github.com/detro/spelunk/v2/types"
	kongcompletion "github.com/jotaen/kong-completion"
)

var (
	// Variables overridden on release via `ldflags -X`
	name    = "spelunk"
	version = "dev"
	commit  = "n/a"
	branch  = "n/a"
	date    = "n/a"
	builtBy = "n/a"

	// Mapping of the variables above to Kong
	kongVars = kong.Vars{
		"name":    name,
		"version": version,
		"commit":  commit,
		"branch":  branch,
		"date":    date,
		"builtBy": builtBy,
	}
)

type CLI struct {
	// Commands
	Dig    DigCmd    `cmd:"" default:"withargs" help:"Dig up a secret (default)."`
	Exists ExistsCmd `cmd:""                    help:"Check if a secret Exists."`
	Creds  CredsCmd  `cmd:""                    help:"Check all configured credentials."`

	Completion kongcompletion.Completion `cmd:"" help:"Generate shell completion scripts."`

	// Configuration for sources
	Config Configurators `embed:""`

	loggingArgs `embed:"" group:"Logging:" prefix:"log."`
}

func (c *CLI) AfterApply() error {
	c.SetupDefaultLogger()

	slog.Debug("cli parsed", slog.Any("cli", c))

	return nil
}

func (c *CLI) NewSpelunker(ctx context.Context) (*spelunk.Spelunker, error) {
	// Enable configured sources
	opts, err := c.Config.SpelunkerOptions(ctx)
	if err != nil {
		return nil, err
	}

	// Enable all modifiers
	opts = append(opts,
		jsonpath.WithJSONPath(),
		tomlpath.WithTOMLPath(),
		xpath.WithXPath(),
		yamlpath.WithYAMLPath(),
	)

	return spelunk.NewSpelunker(opts...), nil
}

func (c *CLI) DigUpSecret(ctx context.Context, coordStr string) (string, error) {
	slog.Debug("Resolving secret coordinate", "coord", coordStr)
	coord, err := types.NewSecretCoord(coordStr)
	if err != nil {
		slog.Error("Failed to parse secret coordinate", "err", err, "coord", coordStr)
		return "", err
	}
	slog.Log(ctx, logger.LevelTrace, "Coordinate parsed", "coord", coordStr, "parsed", coord)

	sp, err := c.NewSpelunker(ctx)
	if err != nil {
		slog.Error("Failed to create spelunker", "err", err)
		return "", err
	}
	slog.Log(ctx, logger.LevelTrace, "Spelunker initialized")

	secret, err := sp.DigUp(ctx, coord)
	if err != nil {
		slog.Error("Failed to dig up secret", "err", err, "coord", coordStr)
		return "", err
	}
	return secret, nil
}

func Parse() *kong.Context {
	var app CLI
	parser := kong.Must(&app,
		kong.Name(name),
		kong.Description(`
			Spelunk - Dig up secrets by coordinates.

			Version: ${version}
			Built from: ${branch} / ${commit}
			Built on: ${date}
			Built by: ${builtBy}

			https://github.com/detro/spelunk
		`),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Summary:   true,
			Compact:   true,
			FlagsLast: true,
			Tree:      true,
		}),
		kongVars,
	)

	kongcompletion.Register(parser)

	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)
	ctx.Bind(&app)

	return ctx
}
