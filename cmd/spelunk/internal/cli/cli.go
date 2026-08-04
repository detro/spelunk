package cli

import (
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
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
	loggingArgs `embed:"" group:"Logging:" prefix:"log."`

	Dig        DigCmd                    `cmd:"" default:"withargs" help:"Dig up a secret."`
	Completion kongcompletion.Completion `cmd:""                    help:"Generate shell completion scripts."`
}

func (c *CLI) AfterApply() error {
	c.SetupDefaultLogger()

	slog.Debug("cli parsed", slog.Any("cli", c))

	return nil
}

func Parse() *kong.Context {
	parser := kong.Must(&CLI{},
		kong.Name(name),
		kong.Description(`
			Spelunk - Dig up secrets by coordinates.
			
			See: https://github.com/detro/spelunk
			Version: ${version}
			Built from: ${branch} / ${commit}
			Built on: ${date}
			Built by: ${builtBy}
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

	return ctx
}
