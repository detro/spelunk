package kubernetes

import (
	"context"
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
	ErrSecretSourceKubernetesInvalidLocation = fmt.Errorf("invalid kubernetes secret location format")
	ErrSecretSourceKubernetesInvalidName     = fmt.Errorf("invalid kubernetes name")

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

func (s *SecretSourceKubernetes) DigUp(ctx context.Context, coord types.SecretCoord) (string, error) {
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
		return "", fmt.Errorf("%w: expected NAMESPACE/NAME/KEY or NAME/KEY, got %q", ErrSecretSourceKubernetesInvalidLocation, coord.Location)
	}

	// Validate
	if !isValidDNSSubdomain(namespace) {
		return "", fmt.Errorf("%w: invalid namespace %q", ErrSecretSourceKubernetesInvalidName, namespace)
	}
	if !isValidDNSSubdomain(name) {
		return "", fmt.Errorf("%w: invalid secret name %q", ErrSecretSourceKubernetesInvalidName, name)
	}
	if len(key) == 0 {
		return "", fmt.Errorf("%w: key cannot be empty", ErrSecretSourceKubernetesInvalidLocation)
	}

	// Retrieve
	secret, err := s.k8sClient.Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return "", fmt.Errorf("%w (%q): %w", types.ErrSecretNotFound, coord.Location, err)
		}
		return "", fmt.Errorf("%w (%q): %w", types.ErrCouldNotFetchSecret, coord.Location, err)
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
