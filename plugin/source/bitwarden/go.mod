module github.com/detro/spelunk/plugin/source/bitwarden/v2

go 1.21

replace github.com/detro/spelunk/v2 => ../../../

replace github.com/detro/spelunk/plugin/modifier/jsonpath/v2 => ../../modifier/jsonpath

require (
	github.com/bitwarden/sdk-go/v2 v2.1.0
	github.com/detro/spelunk/plugin/modifier/jsonpath/v2 v2.1.0
	github.com/detro/spelunk/v2 v2.1.0
	github.com/google/uuid v1.6.0
	github.com/stretchr/testify v1.12.1
)

require (
	github.com/ohler55/ojg v1.28.5 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
)
