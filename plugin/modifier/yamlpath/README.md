# YAMLPath Extractor Modifier (`yp`)

This plugin provides a YAML JSONPath modifier for `spelunk`. It allows you to extract specific values from YAML secrets using JSONPath syntax.

## Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/plugin/modifier/yamlpath"
)

func main() {
	s, _ := spelunk.NewSpelunker(yamlpath.WithYAMLPath())
	
	// Assuming `plain://...` contains `data: { foo: bar }`
	res, _ := s.DigUp(context.Background(), "plain://.../?yp=$.data.foo")
	
	fmt.Println(res) // Outputs: bar
}
```
