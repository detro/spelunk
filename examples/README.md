# Examples

This directory contains runnable examples demonstrating how to use `spelunk` with popular Go configuration and CLI libraries.

Each example is a self-contained Go module. You can `cd` into any directory and run it directly.

## Available Examples

| Example                     | Library Used                                          | Description                                          |
|-----------------------------|-------------------------------------------------------|------------------------------------------------------|
| [basic](./basic/)           | [flag](https://pkg.go.dev/flag)                       | Standard library usage. Simple and dependency-free.  |
| [kong](./kong/)             | [alecthomas/kong](https://github.com/alecthomas/kong) | Declarative command-line parser with struct tagging. |
| [urfave-cli](./urfave-cli/) | [urfave/cli](https://github.com/urfave/cli)           | Popular, feature-rich library for building CLI apps. |
| [viper](./viper/)           | [spf13/viper](https://github.com/spf13/viper)         | Complete configuration solution (files, env, flags). |

## Integration

`spelunk` is designed to be compatible with almost any Go configuration library because `types.SecretCoord` implements `encoding.TextUnmarshaler`.
This allows most libraries to automatically parse configuration strings (e.g., `"k8s://..."`) directly into `SecretCoord` structs.

So, in addition to the examples above, it works with many other libraries.
Check out [Awesome Go: Configuration](https://awesome-go.com/#configuration) if you need more options.
