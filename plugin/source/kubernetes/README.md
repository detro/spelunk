# Kubernetes Secret Source (`k8s://`)

The **Kubernetes** secret source retrieves secrets directly from the Kubernetes API.

## Status

**Plugin**: This source is **opt-in**. It is not enabled by default and requires explicit configuration using `WithKubernetes()`.

## Dependencies

This plugin requires the Kubernetes client libraries:
- `k8s.io/client-go`
- `k8s.io/apimachinery`
- `k8s.io/api`

## Usage

To use the Kubernetes source, use the `k8s://` scheme followed by the namespace (optional), secret name, and key.

### Syntax

**Format 1: With Namespace**
```
k8s://<NAMESPACE>/<SECRET_NAME>/<KEY>
```

**Format 2: Default Namespace**
```
k8s://<SECRET_NAME>/<KEY>
```
*(Defaults to namespace `default`)*

### Examples

Retrieve key `password` from secret `db-creds` in namespace `prod`:
```
k8s://prod/db-creds/password
```

Retrieve key `token` from secret `api-access` in namespace `default`:
```
k8s://api-access/token
```

## Configuration

To use this source, you must initialize `spelunk` with a Kubernetes client:

```go
import (
    "github.com/detro/spelunk"
    "github.com/detro/spelunk/plugin/source/kubernetes"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
)

func main() {
    // 1. Create Kubernetes client
    config, _ := rest.InClusterConfig() // or BuildConfigFromFlags
    clientset, _ := kubernetes.NewForConfig(config)

    // 2. Initialize Spelunker with the Kubernetes plugin
    s := spelunk.NewSpelunker(
        kubernetes.WithKubernetes(clientset.CoreV1()),
    )

    // 3. Dig up secrets
    secret, _ := s.DigUp(ctx, coord)
}
```

## Behavior

1. **Parsing**: Splits the location into Namespace, Name, and Key.
2. **Validation**: Checks if Namespace and Name are valid DNS subdomains (RFC 1123).
3. **Retrieval**: Uses `k8sClient.Secrets(namespace).Get()` to fetch the secret resource.
4. **Extraction**: Looks up the specific `Key` in the secret's `Data` map.
5. **Errors**:
    - Returns `ErrSecretNotFound` if the Secret resource doesn't exist.
    - Returns `ErrSecretKeyNotFound` if the Secret exists but the Key does not.

## Use Cases

- **Kubernetes Operators/Controllers**: Retrieving secrets dynamically without mounting them.
- **In-Cluster Applications**: accessing secrets from other namespaces (if RBAC permits).
