module github.com/detro/spelunk/plugin/modifier/tomlpath/v2

go 1.26.6

replace github.com/detro/spelunk/v2 => ../../../

require (
	github.com/detro/spelunk/v2 v2.1.0
	github.com/oliveagle/jsonpath v0.1.4
	github.com/pelletier/go-toml/v2 v2.4.3
	github.com/stretchr/testify v1.12.1
)

require go.yaml.in/yaml/v3 v3.0.5 // indirect
