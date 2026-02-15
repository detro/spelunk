# (Go) Spelunk

Spelunk is a Golang library for extracting secrets from various sources (Kubernetes, Vault, env vars, files)
using a unified URI-based coordinate system (e.g.,  `k8s://ns/secret/key`).
It simplifies secret management by abstracting backend complexity, providing a consistent API for "digging up"
configuration values in cloud-native apps.

Its primary application (but... you do you) is **command line tools**. Users point at a secret from any _source_:
your tool will adapt based on the `/plugin`s installed.

With a single library, the source of secrets is flexible and adapts to the
environment and the situation.

It (aims to) support(s) the following sources of secret:

| Source of Secrets                     | Type     | Available as | Implemented |
|---------------------------------------|----------|:------------:|:-----------:|
| Environment Variables                 | `env`    |   built-in   |      ✅      |
| File                                  | `file`   |   built-in   |      ✅      |
| Plaintext                             | `plain`  |   built-in   |      ✅      |
| Base64 encoded                        | `base64` |   built-in   |      ⏳      |
| Kubernetes Secrets                    | `k8s`    |   plug-in    |      ✅      |
| Vault                                 | `vault`  |   plug-in    |      ⏳      |
| AWS/GCP/Azure Secrets Manager         | ?        |   plug-in    |      ⏳      |
| AWS/GCP/Azure Keys Management Service | ?        |   plug-in    |      ⏳      |

## Modifiers

Modifiers are _optional behaviour_ applied to a secret after it has been dug-up by Spelunk.

### JSONPath

> [JSONPath] ([RFC 9535]) defines a string syntax for selecting and extracting
> JSON (RFC 8259) values from within a given JSON value.

The [JSONPath] modifier can be used with secrets that are in JSON format.
After parsing, the modifier digs further at the provided path, and returns
the value found there.

> [!WARN]
> The given [JSONPath] is assumed to be referring to a single element.
> Otherwise, returns the first matching.

## License

This project is shared under the [MIT](./LICENSE) license.

[asdf]: https://asdf-vm.com/
[asdf plugins]: https://asdf-vm.com/manage/plugins.html
[task]: https://taskfile.dev/
[task completion]: https://taskfile.dev/docs/installation#setup-completions
[JSONPath]: https://goessner.net/articles/JsonPath/
[RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535
