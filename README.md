# (Go) Spelunk

<img align="right" width="300" src="docs/images/spelunk-logo-transparent.png">

Spelunk is a Golang library for extracting secrets from various sources (Kubernetes, Vault, env vars, files)
using a unified URI-based coordinate system (e.g.,  `k8s://ns/secret/key`).
It simplifies accessing secrets by abstracting a consistent API for "digging up" configuration
values in cloud-native tools and apps.

Its primary application is **command line tools**, but you do you!
Users point at a secret from any _source_: your tool/service/software
will adapt based on the `/plugin`s enabled via the `spelunk.SpelunkerOption`s
provided.

**With a single library, the source of secrets is flexible and adapts to your
environment, situation and/or needs.**

## Sources

Sources are places out of which a secret can be "dug-up".

Some are _built-in_ to `spelunk.Spelunker`, others need to be configured as `spelunk.SpelunkerOption` at creation time:

```go
package main

import (
	"github.com/detro/spelunk"
	"github.com/detro/spelunk/plugin/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Initialize Spelunker with Kubernetes plugin
k8sClient, err := typedcorev1.NewForConfig(restConfig)
// ...
spelunker := spelunk.NewSpelunker(
	kubernetes.WithKubernetes(k8sClient),
)
```

| Source (of Secrets)                   | Type (scheme) |    Is    | Done | Doc           |
|---------------------------------------|---------------|:--------:|:----:|---------------|
| Environment Variables                 | `env://`      | built-in |  ✅   | add link here |
| File                                  | `file://`     | built-in |  ✅   | add link here |
| Plaintext                             | `plain://`    | built-in |  ✅   | add link here |
| Base64 encoded                        | `base64://`   | built-in |  ✅   | add link here |
| Kubernetes Secrets                    | `k8s://`      | plug-in  |  ✅   | add link here |
| Vault                                 | `vault://`    | plug-in  |  ⏳   | ⏳             |
| AWS/GCP/Azure Secrets Manager         | ?             | plug-in  |  ⏳   | ⏳             |
| AWS/GCP/Azure Keys Management Service | ?             | plug-in  |  ⏳   | ⏳             |

## Modifiers

Modifiers are _optional behaviour_ applied to a secret after it has been dug-up by Spelunk.

| Modifier (of Secrets) | Type (query)     |    Is    | Done | Doc           |
|-----------------------|------------------|:--------:|:----:|---------------|
| JSONPath              | `?jp=<JSONPath>` | built-in |  ✅   | add link here |
| XPath                 | `?xp=<XPath>`    | plug-in  |  ⏳   | ⏳             |

## License

This project is shared under the [MIT](./LICENSE) license.

[asdf]: https://asdf-vm.com/
[asdf plugins]: https://asdf-vm.com/manage/plugins.html
[task]: https://taskfile.dev/
[task completion]: https://taskfile.dev/docs/installation#setup-completions
[JSONPath]: https://goessner.net/articles/JsonPath/
[RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535
