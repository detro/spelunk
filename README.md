# (Go) Spelunk

<img align="right" width="300" src="docs/images/spelunk-logo-transparent.png">

Spelunk is a Golang library for extracting secrets from various sources (Kubernetes, Vault, env vars, files)
using a unified URI-based coordinate system (e.g.,  `k8s://ns/secret/key`).
It simplifies accessing secrets by abstracting a consistent API for "digging up" configuration
values in cloud-native tools and apps.

Its primary application is **command line tools**, but... _you do you!_
Users point at a secret from any _source_: your tool/service/software
will adapt based on the _plugins_ enabled via the `spelunk.SpelunkerOption`s
provided.

**With a single library, the source of secrets is flexible and adapts to your
environment, situation and/or needs.**

## Get started

Get the library:

```shell
# Pull the main library
go get github.com/detro/spelunk
# Pull optional plugins
go get github.com/detro/spelunk/plugin/source/kubernetes
```

Setup a new `spelunk.Spelunker` and start digging up secrets:

```golang
package main

import (
	"github.com/detro/spelunk"
	"github.com/detro/spelunk/plugin/source/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Initialize the Kubernetes client...
k8sClient, err := v1.NewForConfig(restConfig)

// Create a Spelunker
spelunker := spelunk.NewSpelunker(
	kubernetes.WithKubernetes(k8sClient),
)

// Get coordinates to a secret from one of many supported sources:
// from Kubernetes... 
coord, err := types.NewSecretCoord("k8s://my-namespace/my-secret/my-data-key")
// ... or from plain text (please don't!)
coord, err := types.NewSecretCoord("plain://MY_PLAINTEXT_SECRET")
// ... or from a local file
coord, err := types.NewSecretCoord("file://secrets.json?jp=$.kafka.password")
// ... or from environment variable
coord, err := types.NewSecretCoord("env://GITHUB_PRIVATE_TOKEN")

// Dig-up secrets!
secret, _ := spelunker.DigUp(ctx, coord)
```

## Sources

Sources are places out of which a secret can be "dug-up".
Some are _built-in_ to `spelunk.Spelunker`, others are _plug-in_ and need to be enabled.

| Source (of Secrets)                   | Type (scheme) |    Is    | Done | Doc                                 |
|---------------------------------------|---------------|:--------:|:----:|-------------------------------------|
| Environment Variables                 | `env://`      | built-in |  ✅   | [`/builtin/source/env`](TODO)       |
| File                                  | `file://`     | built-in |  ✅   | [`/builtin/source/file`](TODO)      |
| Plaintext                             | `plain://`    | built-in |  ✅   | [`/builtin/source/plain`](TODO)     |
| Base64 encoded                        | `base64://`   | built-in |  ✅   | [`/builtin/source/base64`](TODO)    |
| Kubernetes Secrets                    | `k8s://`      | plug-in  |  ✅   | [`/plugin/source/kubernetes`](TODO) |
| Vault                                 | `vault://`    | plug-in  |  ⏳   | ⏳                                   |
| AWS/GCP/Azure Secrets Manager         | ?             | plug-in  |  ⏳   | ⏳                                   |
| AWS/GCP/Azure Keys Management Service | ?             | plug-in  |  ⏳   | ⏳                                   |

## Modifiers

Modifiers are _optional behaviour_ applied to a secret after it has been dug-up by Spelunk.

| Modifier (of Secrets) | Type (query)     |    Is    | Done | Doc                                  |
|-----------------------|------------------|:--------:|:----:|--------------------------------------|
| JSONPath              | `?jp=<JSONPath>` | built-in |  ✅   | [`/builtin/modifier/jsonpath`](TODO) |
| XPath                 | `?xp=<XPath>`    | plug-in  |  ⏳   | ⏳                                    |

## Contributing

If you are interested in contributing (for example, you have a brilliant idea for a plug-in),
we have some [contribution guidelines](./CONTRIBUTING.md).

## License

This project is shared under the [MIT](./LICENSE) license.
