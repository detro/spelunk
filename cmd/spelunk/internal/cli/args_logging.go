package cli

import (
	"github.com/detro/spelunk/cmd/spelunk/internal/logger"
)

const (
	// defaultLogVerbosity is the verbosity passed to `logger.SetupDefaultLogger` when no verbosity flags
	// are specified by the user on the command line.
	defaultLogVerbosity = 1
)

type loggingArgs struct {
	Verbosity int8             `name:"verbose" short:"v" type:"counter" default:"0"      help:"Repeat to increase logging Verbosity"`
	Quietness int8             `name:"quiet"   short:"q" type:"counter" default:"0"      help:"Repeat to decrease logging Verbosity"`
	Format    logger.LogFormat `name:"format"  short:"l"                default:"tinted" help:"Log format (one of: ${enum})"         enum:"tinted,json,boring"`
}

func (l *loggingArgs) SetupDefaultLogger() {
	logger.SetupDefaultLogger(defaultLogVerbosity+l.Verbosity-l.Quietness, l.Format)
}
