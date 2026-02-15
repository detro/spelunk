package kubernetes_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/detro/spelunk"
	"github.com/detro/spelunk/plugin/kubernetes"
	"github.com/detro/spelunk/types"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func TestSecretSourceKubernetes_Type(t *testing.T) {
	s := &kubernetes.SecretSourceKubernetes{}
	require.Equal(t, "k8s", s.Type())
}

const (
	secretNamespace = "test-ns"
	secretName      = "my-secret"
	secretKey       = "password"
	secretValue     = "super-secret-value"
)

func TestSecretSourceKubernetes_DigUp_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	err, k8sClient := setupK3STestContainer(t, ctx)

	// Create namespace
	_, err = k8sClient.Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretNamespace,
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	// Create secret
	_, err = k8sClient.Secrets(secretNamespace).Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			secretKey: []byte(secretValue),
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	// Create secret in default namespace
	_, err = k8sClient.Secrets("default").Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secretName},
		Data:       map[string][]byte{secretKey: []byte(secretValue)},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	// Initialize Spelunker with Kubernetes plugin
	spelunker := spelunk.NewSpelunker(kubernetes.WithKubernetes(k8sClient))

	tests := []struct {
		name     string
		coordStr string
		want     string
		errMatch error
	}{
		{
			name:     "valid secret",
			coordStr: fmt.Sprintf("k8s://%s/%s/%s", secretNamespace, secretName, secretKey),
			want:     secretValue,
		},
		{
			name:     "valid secret in default namespace",
			coordStr: fmt.Sprintf("k8s://%s/%s", secretName, secretKey),
			want:     secretValue,
		},
		{
			name:     "secret not found",
			coordStr: fmt.Sprintf("k8s://%s/missing-secret/%s", secretNamespace, secretKey),
			errMatch: types.ErrSecretNotFound,
		},
		{
			name:     "key not found",
			coordStr: fmt.Sprintf("k8s://%s/%s/missing-key", secretNamespace, secretName),
			errMatch: types.ErrSecretKeyNotFound,
		},
		{
			name:     "invalid namespace name",
			coordStr: "k8s://INVALID-NAMESPACE/secret/key",
			errMatch: kubernetes.ErrSecretSourceKubernetesInvalidName,
		},
		{
			name:     "invalid secret name",
			coordStr: "k8s://ns/INVALID-SECRET/key",
			errMatch: kubernetes.ErrSecretSourceKubernetesInvalidName,
		},
		{
			name:     "invalid location (missing key)",
			coordStr: "k8s://ns/secret/",
			errMatch: kubernetes.ErrSecretSourceKubernetesInvalidLocation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := types.NewSecretCoord(tt.coordStr)
			require.NoError(t, err)

			got, err := spelunker.DigUp(ctx, coord)
			if tt.errMatch != nil {
				require.ErrorIs(t, err, tt.errMatch)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.want, got)
		})
	}
}

func setupK3STestContainer(t *testing.T, ctx context.Context) (error, *typedcorev1.CoreV1Client) {
	k3sContainer, err := k3s.Run(ctx, "rancher/k3s:v1.30.2-k3s1")
	testcontainers.CleanupContainer(t, k3sContainer)
	require.NoError(t, err)

	kubeConfigYaml, err := k3sContainer.GetKubeConfig(ctx)
	require.NoError(t, err)

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeConfigYaml)
	require.NoError(t, err)

	k8sClient, err := typedcorev1.NewForConfig(restConfig)
	require.NoError(t, err)

	return err, k8sClient
}
