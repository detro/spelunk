module github.com/detro/spelunk/cmd/spelunk

go 1.26.6

replace github.com/detro/spelunk/v2 => ../../

require (
	github.com/alecthomas/kong v1.16.1
	github.com/detro/spelunk/v2 v2.0.0
	github.com/jotaen/kong-completion v0.0.14
	github.com/lmittmann/tint v1.2.0
)

require (
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/posener/complete v1.2.3 // indirect
	github.com/riywo/loginshell v0.0.0-20200815045211-7d26008be1ab // indirect
)
