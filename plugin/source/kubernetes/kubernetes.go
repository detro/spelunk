package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/types"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var (
	ErrSecretSourceKubernetesInvalidLocation = fmt.Errorf(
		"invalid Kubernetes secret location format",
	)
	ErrSecretSourceKubernetesInvalidName = fmt.Errorf("invalid Kubernetes name")

	// dnsSubdomainRegex Matches DNS subdomain names as defined
	// in RFC-1123 (https://datatracker.ietf.org/doc/html/rfc1123).
	dnsSubdomainRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`)
)

const (
	defaultNamespace = "default"
)

// SecretSourceKubernetes digs up secrets from Kubernetes Secrets.
// The URI scheme for this source is "k8s".
//
//	k8s://NAMESPACE/NAME/KEY
//	k8s://NAME/KEY (where NAMESPACE is "default")
//	k8s://NAMESPACE/NAME/
//	k8s://NAME/ (where NAMESPACE is "default")
//
// When `/KEY` is appended, Spelunk extracts the specific value in the secret's data map.
// Otherwise, if it ends with `/`, it returns the whole secret's data key-value map as JSON.
//
// This types.SecretSource is a plug-in to spelunker.Spelunker and must be enabled explicitly.
type SecretSourceKubernetes struct {
	k8sClient corev1.SecretsGetter
}

// WithKubernetes enables the SecretSourceKubernetes.
func WithKubernetes(k8sClient corev1.SecretsGetter) spelunk.SpelunkerOption {
	source := &SecretSourceKubernetes{
		k8sClient,
	}
	return spelunk.WithSource(source)
}

var _ types.SecretSource = (*SecretSourceKubernetes)(nil)

func (s *SecretSourceKubernetes) Type() string {
	return "k8s"
}

func (s *SecretSourceKubernetes) DigUp(
	ctx context.Context,
	coord types.SecretCoord,
) (string, error) {
	parts := strings.Split(coord.Location, "/")

	// Take Location apart
	var namespace, name, key string
	switch len(parts) {
	case 2:
		namespace = defaultNamespace
		name = parts[0]
		key = parts[1]
	case 3:
		namespace = parts[0]
		name = parts[1]
		key = parts[2]
	default:
		return "", fmt.Errorf(
			"%w: expected NAMESPACE/NAME/KEY, NAME/KEY, NAMESPACE/NAME/ or NAME/, got %q",
			ErrSecretSourceKubernetesInvalidLocation,
			coord.Location,
		)
	}

	// Validate
	if !isValidDNSSubdomain(namespace) {
		return "", fmt.Errorf(
			"%w: invalid namespace %q",
			ErrSecretSourceKubernetesInvalidName,
			namespace,
		)
	}
	if !isValidDNSSubdomain(name) {
		return "", fmt.Errorf(
			"%w: invalid secret name %q",
			ErrSecretSourceKubernetesInvalidName,
			name,
		)
	}
	// Retrieve
	secret, err := s.k8sClient.Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return "", fmt.Errorf("%w (%q): %w", types.ErrSecretNotFound, coord.Location, err)
		}
		return "", fmt.Errorf("%w (%q): %w", types.ErrCouldNotFetchSecret, coord.Location, err)
	}

	// No key requested: return the whole `Data` map
	if len(key) == 0 {
		// Convert map[string][]byte to map[string]string for JSON serialization
		stringData := make(map[string]string, len(secret.Data))
		for k, v := range secret.Data {
			stringData[k] = string(v)
		}
		dataJsonBytes, err := json.Marshal(stringData)
		if err != nil {
			return "", err
		}
		return string(dataJsonBytes), nil
	}

	if val, found := secret.Data[key]; found {
		return string(val), nil
	}

	return "", fmt.Errorf("%w (%q)", types.ErrSecretKeyNotFound, coord.Location)
}

func isValidDNSSubdomain(s string) bool {
	if len(s) == 0 || len(s) > 253 {
		return false
	}
	return dnsSubdomainRegex.MatchString(s)
}
