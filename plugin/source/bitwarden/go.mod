module github.com/detro/spelunk/plugin/source/bitwarden/v2

go 1.26.5

replace github.com/detro/spelunk/v2 => ../../../

replace github.com/detro/spelunk/plugin/modifier/jsonpath/v2 => ../../modifier/jsonpath

require (
	github.com/bitwarden/sdk-go/v2 v2.1.0
	github.com/detro/spelunk/plugin/modifier/jsonpath/v2 v2.0.0
	github.com/detro/spelunk/v2 v2.0.0
	github.com/google/uuid v1.6.0
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/oliveagle/jsonpath v0.1.4 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
