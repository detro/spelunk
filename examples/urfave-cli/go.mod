module github.com/detro/spelunk/examples/urfave-cli

go 1.26.3

replace github.com/detro/spelunk/v2 => ../../

require (
	github.com/detro/spelunk/v2 v2.0.0-00010101000000-000000000000
	github.com/urfave/cli/v2 v2.27.7
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342 // indirect
)
