# XPath Extractor Modifier (`xp`)

This plugin provides an XPath modifier for `spelunk`. It allows you to extract specific values from XML secrets.

## Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/detro/spelunk/v2"
	"github.com/detro/spelunk/plugin/modifier/xpath/v2"
)

func main() {
	s, _ := spelunk.NewSpelunker(xpath.WithXPath())
	
	// Assuming `plain://...` contains `<data><foo>bar</foo></data>`
	res, _ := s.DigUp(context.Background(), "plain://.../?xp=//foo")
	
	fmt.Println(res) // Outputs: bar
}
```
