module github.com/detro/spelunk/plugin/source/keeper/v2

go 1.26.5

replace github.com/detro/spelunk/v2 => ../../../

replace github.com/detro/spelunk/plugin/modifier/jsonpath/v2 => ../../modifier/jsonpath

require (
	github.com/detro/spelunk/plugin/modifier/jsonpath/v2 v2.0.0
	github.com/detro/spelunk/v2 v2.0.0
	github.com/keeper-security/secrets-manager-go/core v1.7.0
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/oliveagle/jsonpath v0.1.4 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
