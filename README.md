# (Go) Spelunk

<img align="right" width="300" src="docs/images/spelunk-logo-transparent.png">

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

## Get started

Add the library to your project:

```shell
# Pull the main library
go get github.com/detro/spelunk
# Pull optional plugins
go get github.com/detro/spelunk/plugin/source/kubernetes
```

Setup a new `Spelunker` and start digging up secrets:

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

### Examples

Find some useful [`/examples`](./examples) directory for how to use `spelunk` with various
libraries for configuration or command line arguments parsing.

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

import "github.com/detro/spelunk"

type CLI struct {
	Password spelunk.SecretCoord `name:"password" short:"p" help:"your password"`
	// ...
}
```

### Sources (`SecretSource`)

Sources are places out of which a secret can be "dug-up".
Some are _built-in_ to `spelunk.Spelunker`, others are _plug-in_ and need to be enabled.

| Source (of Secrets)                   | Type (scheme) |    Is    | Done |
|---------------------------------------|---------------|:--------:|:----:|
| Environment Variables                 | `env://`      | built-in |  ✅   |
| File                                  | `file://`     | built-in |  ✅   |
| Plaintext                             | `plain://`    | built-in |  ✅   |
| Base64 encoded                        | `base64://`   | built-in |  ✅   |
| Kubernetes Secrets                    | `k8s://`      | plug-in  |  ✅   |
| Vault                                 | `vault://`    | plug-in  |  ⏳   |
| AWS/GCP/Azure Secrets Manager         | ?             | plug-in  |  ⏳   |
| AWS/GCP/Azure Keys Management Service | ?             | plug-in  |  ⏳   |

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

| Modifier (of Secrets) | Type (query)     |    Is    | Done |
|-----------------------|------------------|:--------:|:----:|
| JSONPath extractor    | `?jp=<JSONPath>` | built-in |  ✅   |
| XPath extractor       | `?xp=<XPath>`    | plug-in  |  ⏳   |
| Base64 encoder        | `?base64`        | built-in |  ⏳   |

## Contributing

If you are interested in contributing (for example, you have a brilliant idea for a plug-in),
we have some [contribution guidelines](./CONTRIBUTING.md).

## License

This project is shared under the [MIT](./LICENSE) license.

## Links

* [Architecture documentation](./ARCHITECTURE.md): understand how Spelunk works internally
* [Contribution guidelines](./CONTRIBUTING.md): setting some ground rules
* [Agents documentation](./AGENTS.md): helps LLM-agent augmented developers in their contribution journey
