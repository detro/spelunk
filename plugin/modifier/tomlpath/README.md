# TOMLPath Extractor Modifier (`tp`)

This plugin provides a TOML JSONPath modifier for `spelunk`. It allows you to extract specific values from TOML secrets using JSONPath syntax.

## Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/plugin/modifier/tomlpath"
)

func main() {
	s, _ := spelunk.NewSpelunker(tomlpath.WithTOMLPath())
	
	// Assuming `plain://...` contains `[data]\nfoo = "bar"`
	res, _ := s.DigUp(context.Background(), "plain://.../?tp=$.data.foo")
	
	fmt.Println(res) // Outputs: bar
}
```
