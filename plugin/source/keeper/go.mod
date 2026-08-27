module github.com/detro/spelunk/plugin/source/keeper/v2

go 1.26.6

replace github.com/detro/spelunk/v2 => ../../../

replace github.com/detro/spelunk/plugin/modifier/jsonpath/v2 => ../../modifier/jsonpath

require (
	github.com/detro/spelunk/plugin/modifier/jsonpath/v2 v2.1.0
	github.com/detro/spelunk/v2 v2.1.0
	github.com/keeper-security/secrets-manager-go/core v1.7.0
	github.com/stretchr/testify v1.12.1
)

require (
	github.com/ohler55/ojg v1.28.5 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
)
