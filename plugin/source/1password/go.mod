module github.com/detro/spelunk/plugin/source/1password/v2

go 1.25.0

replace github.com/detro/spelunk/v2 => ../../../

replace github.com/detro/spelunk/plugin/modifier/jsonpath/v2 => ../../modifier/jsonpath

require (
	github.com/1password/onepassword-sdk-go v0.4.1
	github.com/detro/spelunk/plugin/modifier/jsonpath/v2 v2.1.0
	github.com/detro/spelunk/v2 v2.1.0
	github.com/stretchr/testify v1.12.1
)

require (
	github.com/dylibso/observe-sdk/go v0.0.0-20240828172851-9145d8ad07e1 // indirect
	github.com/extism/go-sdk v1.7.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/ianlancetaylor/demangle v0.0.0-20260724033716-83e58baca724 // indirect
	github.com/ohler55/ojg v1.28.5 // indirect
	github.com/tetratelabs/wabin v0.0.0-20230304001439-f6f874872834 // indirect
	github.com/tetratelabs/wazero v1.12.0 // indirect
	go.opentelemetry.io/proto/otlp v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/sys v0.47.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)
