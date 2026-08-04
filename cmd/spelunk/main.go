package main

import (
	"github.com/detro/spelunk/cmd/spelunk/internal/cli"
)

func main() {
	kongContext := cli.Parse()

	kongContext.FatalIfErrorf(kongContext.Run())
}
