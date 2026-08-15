# (Go) Spelunk

<img align="right" width="300" src="docs/images/spelunk-logo-transparent.png">

[![CI](https://github.com/detro/spelunk/actions/workflows/ci.yaml/badge.svg)](https://github.com/detro/spelunk/actions/workflows/ci.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/detro/spelunk/v2.svg)](https://pkg.go.dev/github.com/detro/spelunk/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/detro/spelunk/v2)](https://goreportcard.com/report/github.com/detro/spelunk/v2)
[![License](https://img.shields.io/github/license/detro/spelunk)](LICENSE)
[![Release](https://img.shields.io/github/v/release/detro/spelunk)](https://github.com/detro/spelunk/releases)

**Spelunk** is a Golang library for extracting secrets from various sources
(Kubernetes, Vault, env vars, files) using a unified URI-based string we are calling
**Secret Coordinates**. Here are some example of coordinates:

```shell
# Secret from the namespace `ns`,
# stored inside the secret `my-team-secret`
# at data key `the-key`
k8s://ns/my-team-secret/the-key

# Secret provided in the form
# of base64-encoded string
base64://bXktYmlnLXNlY3JldAo=

# Secret stored in a JSON
# file at a specific field
file://kafka-credentials.json?jp=$.kafka.password
```

Spelunk simplifies the access to secrets by just providing the coordinates for "digging up" configuration
values in cloud-native CLI tools and applications.

Its primary application is **command line tools**, but... _you do you!_
Users point at a secret from any _source_, providing the right _coordinates_:
your tool/service/software can use Spelunk to adapt dynamically and fetch the secret.

**With a single library, the source of secrets is flexible and adapts to your
environment, situation and/or needs.**

Spelunk can be configured to support more [Sources](#sources-secretsource), and users can apply
[Modifiers](#modifiers-secretmodifier) to "prepare" the secret in the exact way they need it.

And, if you want to use Spelunk in your scripts, it also comes with [its own `spelunk` CLI](#spelunk-cli).

## Multi-Module Architecture (since 2.x)

Starting with version `v2.x`, Spelunk implements a highly efficient **Go Multi-Module Workspace** architecture.

Previously, importing Spelunk pulled down every single heavy SDK dependency
(including the AWS, GCP, Azure, HashiCorp Vault, Kubernetes, and 1Password SDKs)
regardless of whether you used them or not. 

From `v2.0.0` onwards:

- **Ultra-Lean Core:** The root core module `github.com/detro/spelunk/v2` is completely bare
  and carries virtually zero production dependencies.
- **Pay Only For What You Use:** Sibling plugins are completely decoupled into isolated submodules.
  Heavyweight dependencies are **only** pulled down by Go if you explicitly choose to import
  and register their corresponding plugin.

## Get started

Add the core library to your project:

```shell
# Pull the ultra-lean core library
go get github.com/detro/spelunk/v2

# Pull only the specific plugins you want to use
go get github.com/detro/spelunk/plugin/source/kubernetes/v2
```

Setup a new `Spelunker` and start digging up secrets:

```golang
package main

import (
	"context"
	"github.com/detro/spelunk/v2"
	"github.com/detro/spelunk/v2/types"
	"github.com/detro/spelunk/plugin/source/kubernetes/v2"
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

### Examples

Find some useful [`/examples`](./examples) directory for how to use `spelunk` with various
libraries for configuration or command line arguments parsing.

### `built-in` vs `plug-in`

Spelunk comes with a bunch of features: [Sources](#sources-secretsource)
and [Modifiers](#modifiers-secretmodifier), the role of which is explained below.
Some are _built-in_, and a new `spelunk.Spelunker` instance comes with those enabled;
others are _plug-in_, and you will have to enable them as `SpelunkerOption`
provided at construction time:

```go
package main

import (
	"github.com/detro/spelunk/v2"
	"github.com/detro/spelunk/plugin/source/kubernetes/v2"
	"github.com/detro/spelunk/plugin/source/vault/v2"
)

_ = spelunk.NewSpelunker(
	kubernetes.WithKubernetes(k8sClient),
	vault.WithVault(vaultClient),
)
```

## Spelunk CLI

Spelunk includes an official standalone command-line interface located in [`cmd/spelunk`](./cmd/spelunk).
It bundles all built-in and plugin backends into a single binary, supporting secret retrieval (`dig`),
existence checks (`exists`), credential verification (`creds`),
and shell auto-completion directly from terminal or CI/CD pipelines.

See [Spelunk CLI Documentation](./cmd/spelunk/README.md) for details and installation instructions.

## Key Types

`spelunk.Spelunker` is the entry point type, and it does its job using
the following types.

### Coordinates (`SecretCoord`)

This is the starting point: take a string containing **Secret Coordinates** as documented
above, and use `types.NewSecretCoord` to turn it into a `SecretCoord`.

This is a generic, secret-type-agnostic representation of how to find a secret. And
it's all that `Spelunker` needs to _dig-up_ the secret.

#### From user input to `SecretCoord`

`SecretCoord` implements `encoding.TextUnmarshaler`, so it can be created through the unmarshalling
of command-line user input, through `json.Unmarshal` and any other type-aware process.

For example, when using the _awesome_ [Kong](https://github.com/alecthomas/kong) library:

```go
package main

import "github.com/detro/spelunk/v2"

type CLI struct {
	Password spelunk.SecretCoord `name:"password" short:"p" help:"your password"`
	// ...
}
```

### Sources (`SecretSource`)

Sources are places out of which a secret can be "dug-up".
Some are _built-in_ to `spelunk.Spelunker`, others are _plug-in_ and need to be enabled.

| Source (of Secrets)                                                              | Type (scheme) | Available as | Status |                                        Doc                                        |
|----------------------------------------------------------------------------------|---------------|:------------:|:------:|:---------------------------------------------------------------------------------:|
| Environment Variables                                                            | `env://`      |   built-in   |   ✅    |    [link](https://pkg.go.dev/github.com/detro/spelunk/v2@main/builtin/source/env)    |
| File                                                                             | `file://`     |   built-in   |   ✅    |   [link](https://pkg.go.dev/github.com/detro/spelunk/v2@main/builtin/source/file)    |
| Plaintext                                                                        | `plain://`    |   built-in   |   ✅    |   [link](https://pkg.go.dev/github.com/detro/spelunk/v2@main/builtin/source/plain)   |
| Base64 encoded                                                                   | `base64://`   |   built-in   |   ✅    |  [link](https://pkg.go.dev/github.com/detro/spelunk/v2@main/builtin/source/base64)   |
| [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)  | `k8s://`      |   plug-in    |   ✅    | [link](https://pkg.go.dev/github.com/detro/spelunk/plugin/source/kubernetes/v2) |
| [Vault](https://www.hashicorp.com/en/products/vault)                             | `vault://`    |   plug-in    |   ✅    |   [link](https://pkg.go.dev/github.com/detro/spelunk/plugin/source/vault/v2)    |
| [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/)                   | `aws://`      |   plug-in    |   ✅    |    [link](https://pkg.go.dev/github.com/detro/spelunk/plugin/source/aws/v2)     |
| [GCP Secrets Manager](https://cloud.google.com/security/products/secret-manager) | `gcp://`      |   plug-in    |   ✅    |    [link](https://pkg.go.dev/github.com/detro/spelunk/plugin/source/gcp/v2)     |
| [Azure Key Vault](https://azure.microsoft.com/en-gb/products/key-vault/)         | `az://`       |   plug-in    |   ✅    |   [link](https://pkg.go.dev/github.com/detro/spelunk/plugin/source/azure/v2)    |
| [1Password](https://developer.1password.com/docs/cli/)                           | `op://`       |   plug-in    |   ✅    | [link](https://pkg.go.dev/github.com/detro/spelunk/plugin/source/1password/v2)  |
| [Bitwarden](https://bitwarden.com/help/cli/)                                     | `bw://`       |   plug-in    | 👷[^1] | [link](https://pkg.go.dev/github.com/detro/spelunk/plugin/source/bitwarden/v2)  |
| [Keeper](https://docs.keeper.io/en/enterprise-guide/commander-cli)               | `kp://`       |   plug-in    | 👷[^1] |   [link](https://pkg.go.dev/github.com/detro/spelunk/plugin/source/keeper/v2)   |
| [LastPass](https://github.com/lastpass/lastpass-cli)                             | `lp://`       |   plug-in    | ❌ [^2] |                                                                                   |
| [Dashlane](https://cli.dashlane.com/)                                            | `dl://`       |   plug-in    | ❌ [^2] |                                                                                   |

[^1]: **Untested**: Looking for contributors with access to a test account/vault!
[^2]: **Not Implemented**: Not implemented due to the lack of maintained Go SDK, no suitable REST API, and no local Testcontainers for simulation.

## Modifiers (`SecretModifier`)

Modifiers are _optional behaviour_ applied to a secret after it has been dug-up by Spelunk.
It can be seen as a _function in the mathematical sense_:

$$
Modifier(SecretVal, Input) = ModifiedSecVal
$$

Each modifier is **applied in the same order provided** in the secret coordinates:

```text
<type>://<location>?mod1=A&mod2=B&mod1=C
```

will result in this sequence:

* `Spelunker` digs-up the secret `<value>` of type `<type>` from the `<location>`
* `mod1` takes the `<value>` and applies `mod1(<value>, A) = <value_A>`
* `mod2` takes the `<value_A>` and applies `mod2(<value_A>, B) = <value_A_B>`
* `mod1` takes the `<value_A_B>` and applies `mod1(<value_A_B>, C) = <value_A_B_C>`
* client code is returned the final `<value_A_B_C>`

| Modifier (of Secrets)             | Type (query)     | Available as | Status |                                           Doc                                            |
|-----------------------------------|------------------|:------------:|:------:|:----------------------------------------------------------------------------------------:|
| Base64 encoder                    | `?b64`           |   built-in   |   ✅    |     [link](https://pkg.go.dev/github.com/detro/spelunk/v2@main/builtin/modifier/base64)     |
| Base64 encoder (alias for `?b64`) | `?b64e`          |   built-in   |   ✅    | [link](https://pkg.go.dev/github.com/detro/spelunk/v2@main/builtin/modifier/base64_encoder) |
| Base64 decoder                    | `?b64d`          |   built-in   |   ✅    | [link](https://pkg.go.dev/github.com/detro/spelunk/v2@main/builtin/modifier/base64_decoder) |
| JSONPath extractor                | `?jp=<JSONPath>` |   plug-in    |   ✅    |    [link](https://pkg.go.dev/github.com/detro/spelunk/plugin/modifier/jsonpath/v2)     |
| XPath extractor                   | `?xp=<XPath>`    |   plug-in    |   ✅    |      [link](https://pkg.go.dev/github.com/detro/spelunk/plugin/modifier/xpath/v2)      |
| YAML JSONPath extractor           | `?yp=<JSONPath>` |   plug-in    |   ✅    |    [link](https://pkg.go.dev/github.com/detro/spelunk/plugin/modifier/yamlpath/v2)     |
| TOML JSONPath extractor           | `?tp=<JSONPath>` |   plug-in    |   ✅    |    [link](https://pkg.go.dev/github.com/detro/spelunk/plugin/modifier/tomlpath/v2)     |
| SHA-2/3 / BLAKE-2/3 / ... hasher  | TBD              |   plug-in    |   ⏳    |                                                                                          |

## Contributing

If you are interested in contributing (for example, you have a brilliant idea for a plug-in),
we have some [contribution guidelines](./CONTRIBUTING.md).

## License

This project is shared under the [MIT](./LICENSE) license.

## Links

* [Spelunk CLI documentation](./cmd/spelunk/README.md): standalone CLI tool and reference
* [Architecture documentation](./ARCHITECTURE.md): understand how Spelunk works internally
* [Contribution guidelines](./CONTRIBUTING.md): setting some ground rules
* [Agents documentation](./AGENTS.md): helps LLM-agent augmented developers in their contribution journey
* [Changelog](./CHANGELOG.md): track the evolution of the project
