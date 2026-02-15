// Package spelunk is a Golang library for extracting secrets from various sources (Kubernetes, Vault, env vars, files)
// using a unified URI-based coordinate system (e.g., k8s://ns/secret/key).
// It simplifies accessing secrets by abstracting a consistent API for "digging up" configuration
// values in cloud-native tools and apps.
//
// Its primary application (but... you do you) is command line tools. Users point at a secret from any source:
// your tool will adapt based on the plugins installed.
//
// With a single library, the source of secrets is flexible and adapts to the
// environment and the situation.
package spelunk
